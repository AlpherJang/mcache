package cache

import (
	"sync"
	"time"

	"github.com/mohae/deepcopy"
)

type Item struct {
	sync.RWMutex
	key        interface{}
	data       interface{}
	createdOn  time.Time
	accessedOn time.Time
	aliveTime  time.Duration
}

func NewCacheItem(key, data interface{}, aliveTime time.Duration) *Item {
	t := time.Now()
	return &Item{
		key:        key,
		data:       data,
		createdOn:  t,
		accessedOn: t,
		aliveTime:  aliveTime,
	}
}

func (item *Item) KeepAlive() {
	item.Lock()
	defer item.Unlock()
	item.accessedOn = time.Now()
}

func (item *Item) updateData(data interface{}) {
	item.Lock()
	defer item.Unlock()
	item.data = data
	t := time.Now()
	item.accessedOn = t
	item.createdOn = t
}

func (item *Item) AccessedOn() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.accessedOn
}

func (item *Item) CreatedOn() time.Time {
	item.RLock()
	defer item.RUnlock()
	return item.createdOn
}

func (item *Item) Key() interface{} {
	item.RLock()
	defer item.RUnlock()
	return item.key
}

func (item *Item) Data() interface{} {
	item.RLock()
	defer item.RUnlock()
	item.accessedOn = time.Now()
	return deepcopy.Copy(item.data)
}
