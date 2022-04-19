package cache

import (
	"errors"
	"sync"
	"time"
)

const (
	DefaultAliveTime = 3 * time.Hour
)

var (
	KeyNotFoundErr = errors.New("key not found")
	KeyExistedErr  = errors.New("exists same key")
	cache          = make(map[string]*Table)
	mutex          sync.RWMutex
)

func Cache(table string) *Table {
	mutex.RLock()
	t, ok := cache[table]
	mutex.RUnlock()
	if !ok {
		mutex.Lock()
		t, ok = cache[table]
		if !ok {
			t = &Table{
				row:  make(map[interface{}]*Item),
				name: table,
			}
			t.cleanupTimer = time.AfterFunc(DefaultAliveTime, t.checkExpire)
			cache[table] = t
		}
		mutex.Unlock()
	}
	return t
}
