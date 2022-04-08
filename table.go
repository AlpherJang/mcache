package cache

import (
	"sync"
	"time"
)

type CacheTable struct {
	name string
	row  map[interface{}]*CacheItem
	sync.RWMutex
	// 预留字段用于清理旧数据
	cleanupTimer    *time.Timer
	cleanupInterval time.Duration
	waitClean       []*CacheItem
	// length
	count int
}

// Count 返回table中的CacheItem数量
func (t *CacheTable) Count() int {
	return t.count
}

// Add 添加CacheItem到table中
func (t *CacheTable) Add(key interface{}, value interface{}) (bool, error) {
	t.Lock()
	defer t.Unlock()
	if _, has := t.row[key]; has {
		return false, KeyExistedErr
	}
	t.row[key] = NewCacheItem(key, value, 0)
	t.count++
	return true, nil
}

// Update 更新table中的CacheItem
func (t *CacheTable) Update(key interface{}, value interface{}) (bool, error) {
	t.Lock()
	defer t.Unlock()
	old, has := t.row[key]
	if !has {
		return false, KeyNotFoundErr
	}
	old.updateData(value)
	return true, nil
}

// UpdateForce 强制更新table中的CacheItem，无论是否存在
func (t *CacheTable) UpdateForce(key interface{}, value interface{}) bool {
	t.Lock()
	defer t.Unlock()
	if old, has := t.row[key]; has {
		old.updateData(value)
	} else {
		t.row[key] = NewCacheItem(key, value, 0)
		t.count++
	}
	return true
}

// Delete 删除table中的CacheItem，若CacheItem不存在则直接return true
func (t *CacheTable) Delete(key interface{}) bool {
	t.Lock()
	defer t.Unlock()
	if _, has := t.row[key]; !has {
		return true
	}
	delete(t.row, key)
	return true
}

// Exists 判断key是否存在于map中
func (t *CacheTable) Exists(key interface{}) bool {
	t.Lock()
	defer t.Unlock()
	_, has := t.row[key]
	return has
}

func (t *CacheTable) Get(key interface{}) (interface{}, error) {
	t.RLock()
	defer t.RUnlock()
	value, has := t.row[key]
	if !has {
		return nil, KeyNotFoundErr
	}
	return value.Data(), nil
}
