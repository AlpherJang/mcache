package cache

import (
	"sync"
	"time"
)

type CacheItem struct {
	sync.RWMutex
	key        interface{}
	data       interface{}
	createdOn  time.Time
	accessedOn time.Time
	aliveTime  time.Duration
}

func NewCacheItem(key, data interface{}, aliveTime time.Duration) *CacheItem {
	t := time.Now()
	return &CacheItem{
		key:        key,
		data:       data,
		createdOn:  t,
		accessedOn: t,
		aliveTime:  aliveTime,
	}
}

func (item *CacheItem) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
}

func (item *CacheItem) updateData(data interface{}) {
	item.Lock()
	defer item.Unlock()
	item.data = data
	t := time.Now()
	item.accessedOn = t
	item.createdOn = t
}

func (item *CacheItem) AccessedOn() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessedOn
}

func (item *CacheItem) CreatedOn() time.Time {
	return item.createdOn
}

func (item *CacheItem) Key() interface{} {
	return item.key
}

func (item *CacheItem) Data() interface{} {
	item.RLock()
	defer item.RUnlock()
	item.accessedOn = time.Now()
	return item.data
}
