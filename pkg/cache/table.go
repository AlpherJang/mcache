package cache

import (
	"context"
	"fmt"
	"github.com/AlpherJang/mcache/pkg/common/errs"
	"github.com/mohae/deepcopy"
	"go.uber.org/zap"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func NewByOption(name string, opts ...Option) *Table {
	t := &Table{
		row:               make(map[interface{}]*Item),
		name:              name,
		cacheAliveTime:    DefaultAliveTime,
		expireCheckPeriod: DefaultAliveTime,

		RWMutex: sync.RWMutex{},
		diskFlushContext: &diskFlushContext{
			msgCh: make(chan notifyMsg, 50),
			done:  make(chan struct{}),
		},
		diskFlushCfg: &diskFlushCfg{
			flushEnabled: false,
			flushHandler: nil,
		},
	}
	for _, o := range opts {
		o(t)
	}
	t.row = make(map[interface{}]*Item, t.cap)
	if t.log == nil {
		zapLogger, _ := zap.NewDevelopment()
		t.log = zapLogger.Sugar().Named(fmt.Sprintf("[cache-table:%s]", name))
	}
	t.cleanupTimer = time.AfterFunc(t.expireCheckPeriod, t.checkExpire)

	if t.flushHandler != nil {
		t.flushEnabled = true
		if t.flushPeriod <= 0 {
			t.flushPeriod = DefaultFlushPeriod
		}
		if t.flushTimeout <= 0 {
			t.flushTimeout = DefaultFlushTimeout
		}
		if t.triggerFlushThreshold <= 0 {
			t.triggerFlushThreshold = DefaultFlushThreshold
		}
		if t.minFlushInterval <= 0 {
			t.minFlushInterval = DefaultMinFlushInterval
		}
		t.diskFlushContext.flushTimer = time.AfterFunc(t.flushPeriod, t.triggerFlush)
	} else {
		t.flushEnabled = false
	}

	go t.watch()
	return t
}

type diskFlushCfg struct {
	flushEnabled                                bool
	flushHandler                                DiskFlushHandler
	flushPeriod, flushTimeout, minFlushInterval time.Duration
	triggerFlushThreshold                       int64
	stopFlushCh                                 <-chan struct{}
	rmTableOnFlushFinished                      bool
}

type diskFlushContext struct {
	msgCh         chan notifyMsg
	counter       int64
	lastFlushTime time.Time
	finished      bool
	lastFlushErr  error
	flushTimer    *time.Timer
	done          chan struct{}
}

type Table struct {
	name string
	cap  int64
	row  map[interface{}]*Item
	sync.RWMutex
	// 预留字段用于清理旧数据
	cacheAliveTime, expireCheckPeriod time.Duration
	cleanupTimer                      *time.Timer
	// length
	count            int
	callbackObj      ObjectEventHandlerFuncs
	diskFlushContext *diskFlushContext
	*diskFlushCfg
	log     LogPrinter
	deleted bool
}

// Count 返回table中的CacheItem数量
func (t *Table) Count() int {
	return t.count
}

// RegisterCallback 注册回调函数，进行增删改查操作时回调对应函数
func (t *Table) RegisterCallback(obj ObjectEventHandlerFuncs) {
	t.callbackObj = obj
}

// Add 添加CacheItem到table中
func (t *Table) Add(key interface{}, value interface{}) (bool, errs.InnerError) {
	t.Lock()
	defer t.Unlock()
	if _, has := t.row[key]; has {
		return false, errs.KeyExistsErr
	}
	t.row[key] = NewCacheItem(key, value, t.cacheAliveTime)
	t.count++
	t.notifyChange(false)
	if t.callbackObj.AddFn != nil {
		t.callbackObj.AddFn(key, deepcopy.Copy(value))
	}
	return true, nil
}

// Update 更新table中的CacheItem
func (t *Table) Update(key interface{}, value interface{}, checkFunc ...UpdateCheckFunc) errs.InnerError {
	// 这里加读锁即可，可以让不同key的update并发操作
	// 对于相同的key，其item的updateData方法会加写锁，也不会产生并发问题
	t.RLock()
	defer t.RUnlock()
	old, has := t.get(key)
	if !has {
		return errs.CacheNotFoundErr
	}

	oldData := old.Data()

	for _, fn := range checkFunc {
		if !fn(oldData) {
			return errs.UpdateCacheErr
		}
	}
	// item的更新操作内部已加写锁
	if old.updateData(value) {
		t.notifyChange(false)
	}
	if t.callbackObj.UpdateFn != nil {
		t.callbackObj.UpdateFn(key, oldData, deepcopy.Copy(value))
	}
	return nil
}

// UpdateWithCheck 更新table中的CacheItem（当存在时，只有通过校验才会更新；不存在则直接add）
func (t *Table) UpdateWithCheck(key interface{}, value interface{}, checkFunc ...UpdateCheckFunc) (bool, errs.InnerError) {
	t.Lock()
	defer t.Unlock()
	if old, has := t.row[key]; has {
		oldData := old.Data()
		for _, fn := range checkFunc {
			if !fn(oldData) {
				return true, errs.UpdateCacheErr
			}
		}
		if old.updateData(value) {
			t.log.Debugf("item changed, oldData=%+v, newData=%+v", oldData, value)
			t.notifyChange(false)
		}
		if t.callbackObj.UpdateFn != nil {
			t.callbackObj.UpdateFn(key, oldData, deepcopy.Copy(value))
		}
		return true, nil
	} else {
		t.row[key] = NewCacheItem(key, value, t.cacheAliveTime)
		t.count++
		t.notifyChange(false)
		if t.callbackObj.AddFn != nil {
			t.callbackObj.AddFn(key, deepcopy.Copy(value))
		}
		return false, nil
	}
}

// UpdateForce 强制更新table中的CacheItem，无论是否存在
func (t *Table) UpdateForce(key interface{}, value interface{}) bool {
	t.Lock()
	defer t.Unlock()
	if old, has := t.get(key); has {
		oldData := old.Data()
		if old.updateData(value) {
			t.notifyChange(false)
		}
		if t.callbackObj.UpdateFn != nil {
			t.callbackObj.UpdateFn(key, oldData, value)
		}
	} else {
		t.row[key] = NewCacheItem(key, value, t.cacheAliveTime)
		t.count++
		t.notifyChange(false)
		if t.callbackObj.AddFn != nil {
			t.callbackObj.AddFn(key, deepcopy.Copy(value))
		}
	}
	return true
}

// Delete 删除table中的CacheItem，若CacheItem不存在则直接return false
func (t *Table) Delete(key interface{}) bool {
	t.Lock()
	defer t.Unlock()
	item, has := t.row[key]
	if !has {
		return false
	}
	t.delete(key, item)
	return true
}

func (t *Table) delete(key interface{}, item *Item) {
	delete(t.row, key)
	t.count--
	t.notifyChange(false)
	if t.callbackObj.RemoveFn != nil {
		t.callbackObj.RemoveFn(key, item.Data())
	}
}

// Exists 判断key是否存在于map中
func (t *Table) Exists(key interface{}) bool {
	t.RLock()
	defer t.RUnlock()
	_, has := t.row[key]
	return has
}

// Get 获取获取指定key的缓存数据
func (t *Table) Get(key interface{}) (interface{}, errs.InnerError) {
	t.RLock()
	defer t.RUnlock()
	if value, has := t.get(key); !has {
		return nil, errs.CacheNotFoundErr
	} else {
		return value.Data(), nil
	}
}

func (t *Table) get(key interface{}) (*Item, bool) {
	value, has := t.row[key]
	return value, has
}

// GetAndDelete 取出key对应的value的同时，从缓存中删除
func (t *Table) GetAndDelete(key interface{}) (interface{}, errs.InnerError) {
	t.Lock()
	defer t.Unlock()
	value, has := t.row[key]
	if !has {
		return nil, errs.CacheNotFoundErr
	} else {
		t.delete(key, value)
		return value.Data(), nil
	}
}

// List 获取缓存中所有的数据
func (t *Table) List() []interface{} {
	t.RLock()
	defer t.RUnlock()
	list := make([]interface{}, 0)
	for _, item := range t.row {
		list = append(list, item.Data())
	}
	return list
}

// RangeByKey 根据key循环遍历数据
func (t *Table) RangeByKey(checkRange ...CheckFunc) map[interface{}]interface{} {
	t.RLock()
	defer t.RUnlock()
	mp := make(map[interface{}]interface{}, t.count)
	for _, item := range t.row {
		if len(checkRange) == 0 || checkRange[0](item.Key()) {
			mp[item.Key()] = item.Data()
		}
	}
	return mp
}

// CloneRows 对rows进行深拷贝
func (t *Table) CloneRows() map[interface{}]interface{} {
	t.RLock()
	defer t.RUnlock()
	mp := make(map[interface{}]interface{}, t.count)
	for _, item := range t.row {
		k, v := item.KvOnly()
		mp[k] = v
	}
	return mp
}

func (t *Table) NewClonedItemRanger() ItemRanger {
	return &tableRanger{
		table: t,
	}
}

func (t *Table) WaitFlushDone(stopCh <-chan struct{}) bool {
	if !t.flushEnabled {
		return true
	}

	select {
	case <-stopCh:
		return false
	case <-t.diskFlushContext.done:
		return true
	}
}

func (t *Table) notifyChange(triggerByTimer bool) {
	if !t.flushEnabled {
		return
	}
	select {
	// 防止并发高时由于消费者来不及消费而卡住
	case t.diskFlushContext.msgCh <- notifyMsg{triggerByTimer: triggerByTimer}:
	default:
	}
}

func (t *Table) watch() {
	if t.stopFlushCh == nil {
		ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		t.stopFlushCh = ctx.Done()
	}

	for {
		select {
		case <-t.stopFlushCh:
			t.log.Debugf("receive stop flush signal, flush once and return")
			t.flush(true)
			return
		case msg := <-t.diskFlushContext.msgCh:
			if !msg.triggerByTimer {
				t.diskFlushContext.counter++
			}
			t.flush(false)
		}
		if t.diskFlushContext.finished {
			t.log.Debugf("flush finished, return")
			close(t.diskFlushContext.done)
			return
		}
	}
}

type tableRanger struct {
	table      *Table
	clonedRows map[interface{}]interface{}
}

func (t *tableRanger) RangeByKey(checkRange ...CheckFunc) (mp map[interface{}]interface{}) {
	if t.clonedRows == nil {
		t.clonedRows = t.table.CloneRows()
	}
	mp = make(map[interface{}]interface{}, len(t.clonedRows))
	for k, v := range t.clonedRows {
		if len(checkRange) == 0 || checkRange[0](k) {
			mp[k] = v
		}
	}
	return mp
}

func (t *tableRanger) Get(key interface{}) (interface{}, bool) {
	if t.clonedRows == nil {
		t.clonedRows = t.table.CloneRows()
	}
	if v, exist := t.clonedRows[key]; exist {
		return v, true
	} else {
		return nil, false
	}
}

func (t *Table) flush(stopNow bool) bool {
	if !t.timeToFlush(stopNow) {
		return t.diskFlushContext.finished
	}
	t.log.Infof("time to flush now")
	ctx, cancel := context.WithTimeout(context.Background(), t.flushTimeout)
	defer cancel()
	finished, err := t.flushHandler(ctx, t.name, t.NewClonedItemRanger())
	if err != nil {
		t.log.Errorf("flush with error: %v", err)
	} else {
		t.log.Debugf("flush success, finished=%t", finished)
	}

	t.diskFlushContext.flushTimer.Reset(t.flushPeriod)

	t.diskFlushContext.lastFlushErr = err

	if stopNow {
		finished = true
	}

	if finished {
		t.diskFlushContext.flushTimer.Stop()
		if t.rmTableOnFlushFinished {
			t.cleanupTimer.Stop()
			t.deleted = true
		}
	}

	t.diskFlushContext.finished = finished

	t.diskFlushContext.counter = 0
	t.diskFlushContext.lastFlushTime = time.Now()
	return t.diskFlushContext.finished
}

func (t *Table) timeToFlush(stopNow bool) bool {
	if !t.flushEnabled {
		return false
	}
	if t.diskFlushContext.finished {
		return false
	}
	if stopNow {
		if t.diskFlushContext.counter > 0 || t.diskFlushContext.lastFlushErr != nil {
			return true
		}
	} else {
		// 以下3种情况都会触发flush
		// 1. counter超过了阈值并且距离上次flush已经过了最小间隔时间
		// 2. counter > 0 并且距离上次flush已经过了最小间隔时间*2
		// 3. 距离上次flush已经过了最小间隔时间*10
		if (t.diskFlushContext.counter >= t.triggerFlushThreshold && time.Since(t.diskFlushContext.lastFlushTime) >= t.minFlushInterval) ||
			(t.diskFlushContext.counter > 0 && time.Since(t.diskFlushContext.lastFlushTime) >= 2*t.minFlushInterval) ||
			time.Since(t.diskFlushContext.lastFlushTime) >= 10*t.minFlushInterval {
			t.log.Debugf("current counter=%d, triggerFlushThreshold=%d", t.diskFlushContext.counter, t.triggerFlushThreshold)
			return true
		}
	}
	return false
}

// CheckExpire 检查缓存是否过期
func (t *Table) checkExpire() {
	if t.deleted {
		return
	}
	t.Lock()
	defer t.Unlock()
	now := time.Now()
	for key, item := range t.row {
		item.RLock()
		aliveTime := item.aliveTime
		accessedOn := item.accessedOn
		item.RUnlock()
		if aliveTime == 0 {
			continue
		}
		if now.Sub(accessedOn) >= aliveTime {
			// 这里的删除无需再加锁了
			t.delete(key, item)
		}
	}
	t.cleanupTimer = time.AfterFunc(t.expireCheckPeriod, t.checkExpire)
}

// triggerFlush 定期触发flush
func (t *Table) triggerFlush() {
	if !t.flushEnabled || t.diskFlushContext.finished || t.deleted {
		return
	}
	t.notifyChange(true)
	t.diskFlushContext.flushTimer = time.AfterFunc(t.flushPeriod, t.triggerFlush)
}
