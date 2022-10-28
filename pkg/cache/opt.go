package cache

import "time"

func WithAliveTime(aliveTime time.Duration) Option {
	return func(t *Table) {
		t.cacheAliveTime = aliveTime
		if t.cacheAliveTime > MinExpireCheckPeriod {
			t.expireCheckPeriod = t.cacheAliveTime
		} else {
			t.expireCheckPeriod = MinExpireCheckPeriod
		}
	}
}

func WithDiskFlushHandler(handler DiskFlushHandler) Option {
	return func(t *Table) {
		if t.diskFlushCfg == nil {
			t.diskFlushCfg = &diskFlushCfg{}
		}
		t.diskFlushCfg.flushHandler = handler
	}
}

func WithDiskFlushPeriod(period time.Duration) Option {
	return func(t *Table) {
		if t.diskFlushCfg == nil {
			t.diskFlushCfg = &diskFlushCfg{}
		}
		t.diskFlushCfg.flushPeriod = period
	}
}

func WithDiskFlushTimeout(timeout time.Duration) Option {
	return func(t *Table) {
		if t.diskFlushCfg == nil {
			t.diskFlushCfg = &diskFlushCfg{}
		}
		t.diskFlushCfg.flushTimeout = timeout
	}
}

func WithMinFlushInterval(interval time.Duration) Option {
	return func(t *Table) {
		if t.diskFlushCfg == nil {
			t.diskFlushCfg = &diskFlushCfg{}
		}
		t.diskFlushCfg.minFlushInterval = interval
	}
}

func WithTriggerFlushThreshold(threshold int64) Option {
	return func(t *Table) {
		if t.diskFlushCfg == nil {
			t.diskFlushCfg = &diskFlushCfg{}
		}
		t.diskFlushCfg.triggerFlushThreshold = threshold
	}
}

func WithStopFlushCh(stopCh <-chan struct{}) Option {
	return func(t *Table) {
		if t.diskFlushCfg == nil {
			t.diskFlushCfg = &diskFlushCfg{}
		}
		t.diskFlushCfg.stopFlushCh = stopCh
	}
}

func WithRmTableOnFlushFinished(rmTableOnFlushFinished bool) Option {
	return func(t *Table) {
		if t.diskFlushCfg == nil {
			t.diskFlushCfg = &diskFlushCfg{}
		}
		t.diskFlushCfg.rmTableOnFlushFinished = rmTableOnFlushFinished
	}
}

func WithLogger(log LogPrinter) Option {
	return func(t *Table) {
		t.log = log
	}
}

func WithCap(cap int64) Option {
	return func(t *Table) {
		if cap >= 0 {
			t.cap = cap
		}
	}
}

func WithCallbackHandlers(handlers ObjectEventHandlerFuncs) Option {
	return func(t *Table) {
		t.callbackObj = handlers
	}
}
