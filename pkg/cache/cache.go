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

func ListTable(filters ...TableFilter) ([]string, errs.InnerError) {
	mutex.RLock()
	defer mutex.RUnlock()
	res := make([]string, 0, len(cache))
	for _, item := range cache {
		res = append(res, item.name)
	}
	return res, nil
}

func DropTable(name string) {
	mutex.Lock()
	defer mutex.Unlock()
	if _, ok := cache[name]; ok {
		delete(cache, name)
	}
	return
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
