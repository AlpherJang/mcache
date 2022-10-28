package cache

import (
	"bytes"
	"encoding/json"
	"github.com/mohae/deepcopy"
	"reflect"
	"sync"
	"time"
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

func (item *Item) updateData(data interface{}) bool {
	item.Lock()
	defer item.Unlock()
	changed := !item.equal(data)
	item.data = data
	t := time.Now()
	item.accessedOn = t
	item.createdOn = t
	return changed
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

func (item *Item) KvOnly() (interface{}, interface{}) {
	item.RLock()
	defer item.RUnlock()
	return deepcopy.Copy(item.key), deepcopy.Copy(item.data)
}

func (item *Item) equal(data interface{}) bool {
	b1, err1 := json.Marshal(data)
	b2, err2 := json.Marshal(item.data)
	if err1 == nil && err2 == nil {
		return bytes.Equal(b1, b2)
	}
	return reflect.DeepEqual(data, item.data)
}
