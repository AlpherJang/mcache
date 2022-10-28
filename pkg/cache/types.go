package cache

import (
	"context"
	"time"
)

const (
	DefaultAliveTime     = 3 * time.Hour
	MinExpireCheckPeriod = 1 * time.Minute

	DefaultFlushPeriod      = 5 * time.Second
	DefaultFlushTimeout     = 5 * time.Second
	DefaultFlushThreshold   = 5
	DefaultMinFlushInterval = time.Second
)

// ObjectEventHandlerFuncs 注册回调函数
type ObjectEventHandlerFuncs struct {
	AddFn    func(key, item interface{})
	UpdateFn func(key, item, newItem interface{})
	RemoveFn func(key, item interface{})
}

type CheckFunc func(key interface{}) bool
type UpdateCheckFunc func(value interface{}) bool

type notifyMsg struct {
	triggerByTimer bool
}
type DiskFlushHandler func(ctx context.Context, name string, itemRanger ItemRanger) (finished bool, err error)

type Option func(table *Table)

type LogPrinter interface {
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
	Errorf(format string, v ...interface{})
}

type ItemRanger interface {
	RangeByKey(checkRange ...CheckFunc) (mp map[interface{}]interface{})
	Get(key interface{}) (interface{}, bool)
}
