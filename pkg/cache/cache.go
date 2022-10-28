package cache

import (
	"github.com/AlpherJang/mcache/pkg/common/errs"
	"sync"
	"time"
)

const cleanPeriod = 10 * time.Minute

var (
	cache = make(map[string]*Table)
	mutex sync.RWMutex
)

func init() {
	// 定时从cache map中清理已被删除的table
	time.AfterFunc(cleanPeriod, cleanDeletedTable)
}

func RemoveTable(table string) {
	mutex.Lock()
	defer mutex.Unlock()
	_, ok := cache[table]
	if ok {
		delete(cache, table)
	}
}

func cleanDeletedTable() {
	mutex.Lock()
	for key, tb := range cache {
		if tb.deleted {
			delete(cache, key)
		}
	}
	mutex.Unlock()
	time.AfterFunc(cleanPeriod, cleanDeletedTable)
}

func TableExists(table string) bool {
	mutex.RLock()
	t, ok := cache[table]
	mutex.RUnlock()
	if ok && t.deleted {
		RemoveTable(table)
		return false
	}
	return ok
}

func WithOptionCache(table string, opts ...Option) *Table {
	mutex.RLock()
	t, ok := cache[table]
	mutex.RUnlock()
	if !ok || t.deleted {
		mutex.Lock()
		t, ok = cache[table]
		if !ok || t.deleted {
			t = NewByOption(table, opts...)
			cache[table] = t
		}
		mutex.Unlock()
	}
	return t
}

// GetTable search table from cache, and return errs.TableNotFoundErr when table not exist
func GetTable(table string) (*Table, errs.InnerError) {
	mutex.RLock()
	t, ok := cache[table]
	mutex.RUnlock()
	if ok {
		if !t.deleted {
			return t, nil
		}
		RemoveTable(table)
	}
	return nil, errs.TableNotFoundErr
}
