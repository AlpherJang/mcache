package cache

import (
	"errors"
	"github.com/AlpherJang/mcache/pkg/common/errs"
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
	TableNotFoundErr       = errors.New("table not found")
	cache                  = make(map[string]*Table)
	mutex                  sync.RWMutex
)

type CheckFunc func(key interface{}) bool
type UpdateCheckFunc func(value interface{}) bool

// GetTable search table from cache, and return TableNotFoundErr when table not exist
func GetTable(table string) (*Table, errs.InnerError) {
	if item, ok := cache[table]; !ok {
		return nil, errs.TableNotFoundErr
	} else {
		return item, nil
	}
}

// Cache create table and save to cache, if table existed, get it and return
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
