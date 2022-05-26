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
	KeyNotFoundErr         = errors.New("key not exist")
	KeyExistedErr          = errors.New("key already exist")
	UpdateCheckRejectedErr = errors.New("update check rejected")
	cache                  = make(map[string]*Table)
	mutex                  sync.RWMutex
)

type CheckFunc func(key interface{}) bool
type UpdateCheckFunc func(value interface{}) bool

func Cache(table string, liveTime time.Duration) *Table {
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
			t.cleanupTimer = time.AfterFunc(liveTime, t.checkExpire)
			cache[table] = t
		}
		mutex.Unlock()
	}
	return t
}
