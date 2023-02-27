package cache

import (
	"sync"
	"time"

	"github.com/AlpherJang/mcache/pkg/common/errs"
)

// ObjectEventHandlerFuncs 注册回调函数
type ObjectEventHandlerFuncs struct {
	AddFn    func(key, item interface{})
	UpdateFn func(key, item, newItem interface{})
	RemoveFn func(key, item interface{})
}

type Table struct {
	name string
	row  map[interface{}]*Item
	sync.RWMutex
	// 预留字段用于清理旧数据
	cleanupTimer    *time.Timer
	cleanupInterval time.Duration
	waitClean       []*Item
	// length
	count       int
	callbackObj ObjectEventHandlerFuncs
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
		return false, errs.AddCacheErr
	}
	t.row[key] = NewCacheItem(key, value, DefaultAliveTime)
	t.count++
	if t.callbackObj.AddFn != nil {
		t.callbackObj.AddFn(key, value)
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

	for _, fn := range checkFunc {
		if !fn(old.Data()) {
			return errs.UpdateCacheErr
		}
	}

	// item的更新操作内部已加写锁
	old.updateData(value)
	if t.callbackObj.UpdateFn != nil {
		t.callbackObj.UpdateFn(key, old.Data(), value)
	}
	return nil
}

// UpdateForce 强制更新table中的CacheItem，无论是否存在
func (t *Table) UpdateForce(key interface{}, value interface{}) bool {
	t.Lock()
	defer t.Unlock()
	if old, has := t.get(key); has {
		old.updateData(value)
		if t.callbackObj.UpdateFn != nil {
			t.callbackObj.UpdateFn(key, old.Data(), value)
		}
	} else {
		t.row[key] = NewCacheItem(key, value, DefaultAliveTime)
		t.count++
		if t.callbackObj.AddFn != nil {
			t.callbackObj.AddFn(key, value)
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
func (t *Table) RangeByKey(checkRange CheckFunc) []interface{} {
	t.RLock()
	defer t.RUnlock()
	list := make([]interface{}, 0)
	for _, item := range t.row {
		if checkRange(item.Key()) {
			list = append(list, item.Data())
		}
	}
	return list
}

// CheckExpire 检查缓存是否过期
func (t *Table) checkExpire() {
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
	t.cleanupTimer = time.AfterFunc(DefaultAliveTime, t.checkExpire)
}
