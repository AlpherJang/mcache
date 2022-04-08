package cache

import (
	"errors"
	"sync"
)

var (
	KeyNotFoundErr = errors.New("key not found")
	KeyExistedErr  = errors.New("exists same key")
	cache          = make(map[string]*CacheTable)
	mutex          sync.RWMutex
)

func Cache(table string) *CacheTable {
	mutex.RLock()
	t, ok := cache[table]
	mutex.RUnlock()
	if !ok {
		mutex.Lock()
		t, ok = cache[table]
		if !ok {
			t = &CacheTable{
				row:  make(map[interface{}]*CacheItem),
				name: table,
			}
			cache[table] = t
		}
		mutex.Unlock()
	}
	return t
}
