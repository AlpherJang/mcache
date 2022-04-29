package cache

import (
	"sync"
	"time"
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
func (t *Table) Add(key interface{}, value interface{}) (bool, error) {
	t.Lock()
	defer t.Unlock()
	if _, has := t.row[key]; has {
		return false, KeyExistedErr
	}
	t.row[key] = NewCacheItem(key, value, DefaultAliveTime)
	t.count++
	if t.callbackObj.AddFn != nil {
		t.callbackObj.AddFn(key, value)
	}
	return true, nil
}

// Update 更新table中的CacheItem
func (t *Table) Update(key interface{}, value interface{}) error {
	t.Lock()
	defer t.Unlock()
	old, has := t.row[key]
	if !has {
		return KeyNotFoundErr
	}
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
	if old, has := t.row[key]; has {
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

// Delete 删除table中的CacheItem，若CacheItem不存在则直接return true
func (t *Table) Delete(key interface{}) bool {
	t.Lock()
	defer t.Unlock()
	item, has := t.row[key]
	if !has {
		return true
	}
	delete(t.row, key)
	if t.callbackObj.RemoveFn != nil {
		t.callbackObj.RemoveFn(key, item.Data())
	}
	return true
}

// Exists 判断key是否存在于map中
func (t *Table) Exists(key interface{}) bool {
	t.Lock()
	defer t.Unlock()
	_, has := t.row[key]
	return has
}

// Get 获取获取指定key的缓存数据
func (t *Table) Get(key interface{}) (interface{}, error) {
	t.RLock()
	defer t.RUnlock()
	value, has := t.row[key]
	if !has {
		return nil, KeyNotFoundErr
	}
	return value.Data(), nil
}

// GetAndDelete 取出key对应的value的同时，从缓存中删除
func (t *Table) GetAndDelete(key interface{}) (interface{}, error) {
	t.Lock()
	defer t.Unlock()
	value, has := t.row[key]
	if !has {
		return nil, KeyNotFoundErr
	} else {
		delete(t.row, key)
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
			t.Delete(key)
		}
	}
	t.cleanupTimer = time.AfterFunc(DefaultAliveTime, t.checkExpire)
}
