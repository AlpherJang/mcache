package cache

import (
	"sync"
	"time"

	"github.com/AlpherJang/mcache/pkg/common/errs"
)

const (
	DefaultAliveTime = 3 * time.Hour
)

var (
	cache = make(map[string]*Table)
	mutex sync.RWMutex
)

const cleanPeriod = 10 * time.Minute

func init() {
	// 定时从cache map中清理已被删除的table
	time.AfterFunc(cleanPeriod, cleanList)
}

type CheckFunc func(key interface{}) bool
type UpdateCheckFunc func(value interface{}) bool

// GetTable search table from cache, and return TableNotFoundErr when table not exist
func GetTable(table string) (*Table, errs.InnerError) {
	mutex.RLock()
	defer mutex.RUnlock()
	if item, ok := cache[table]; !ok || item.deleted {
		return nil, errs.TableNotFoundErr
	} else {
		return item, nil
	}
}

func cleanList() {
	mutex.Lock()
	defer mutex.Unlock()
	for key, table := range cache {
		if table.deleted {
			delete(cache, key)
		}
	}
	time.AfterFunc(cleanPeriod, cleanList)
}

func ListTable(filters ...TableFilter) ([]string, errs.InnerError) {
	mutex.RLock()
	defer mutex.RUnlock()
	res := make([]string, 0, len(cache))
	for _, item := range cache {
		if !item.deleted {
			res = append(res, item.name)
		}
	}
	return res, nil
}

func DropTable(name string) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := cache[name]; ok {
		cache[name].deleted = true
	}
	return
}

// Cache create table and save to cache, if table existed, get it and return
func Cache(table string, liveTime time.Duration) *Table {
	mutex.RLock()
	t, ok := cache[table]
	mutex.RUnlock()
	mutex.Lock()
	if !ok {
		t = newTable(table, liveTime)
		cache[table] = t
	} else if t.deleted {
		delete(cache, table)
		t = newTable(table, liveTime)
		cache[table] = t
	}
	mutex.Unlock()
	return t
}

func newTable(name string, liveTime time.Duration) *Table {
	t := &Table{
		row:  make(map[interface{}]*Item),
		name: name,
	}
	t.cleanupTimer = time.AfterFunc(liveTime, t.checkExpire)
	return t
}
