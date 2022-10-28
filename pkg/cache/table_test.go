package cache

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type testContext struct {
	*testing.T
	expectedSum, actualSum int64
	stopCh                 chan struct{}
	tableName              string
	updateConcurrency      int64
}

func (c *testContext) flush(ctx context.Context, name string, itemRanger ItemRanger) (finished bool, err error) {
	c.Logf("trigger flush...")

	time.Sleep(time.Millisecond * time.Duration(rand.Intn(500)))

	items := itemRanger.RangeByKey()
	c.actualSum = 0
	for key, item := range items {
		c.Logf("item key: %v, value: %v", key, item)
		c.actualSum += item.(int64)
	}

	time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))

	if c.expectedSum <= c.actualSum {
		c.Logf("flush finished")
		return true, nil
	} else {
		c.Logf("current sum=%d, not finished", c.actualSum)
		return false, nil
	}
}

func TestDiskFlush(t *testing.T) {
	rand.Seed(time.Now().Unix())

	c := &testContext{
		T:                 t,
		expectedSum:       0,
		stopCh:            make(chan struct{}),
		tableName:         "table01",
		updateConcurrency: 1000,
	}

	c.expectedSum = c.updateConcurrency * (c.updateConcurrency + 1) / 2

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(2*c.updateConcurrency))
	defer cancel()

	tb := NewByOption(c.tableName, WithDiskFlushHandler(c.flush), WithRmTableOnFlushFinished(true))
	wg := &sync.WaitGroup{}
	wg.Add(int(c.updateConcurrency))
	for i := 0; i < int(c.updateConcurrency); i++ {
		go func(i int) {
			defer wg.Done()
			if i%3 == 0 {
				time.Sleep(time.Millisecond * time.Duration(3000+rand.Intn(2000)))
			} else {
				time.Sleep(time.Millisecond * time.Duration(100+rand.Intn(900)))
			}
			tb.UpdateForce(fmt.Sprintf("key-%d", i+1), int64(i+1))
			if i%3 == 0 {
				time.Sleep(time.Millisecond * time.Duration(1000+rand.Intn(2000)))
			} else {
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(1000)))
			}
		}(i)
	}
	wg.Wait()
	if finished := tb.WaitFlushDone(ctx.Done()); !finished {
		t.Fatalf("wait flush done timeout: %v", ctx.Err())
	}
	if c.actualSum != c.expectedSum {
		t.Fatalf("actualSum=%d, expectedSum=%d, not equal", c.actualSum, c.expectedSum)
	}
	if TableExists(c.tableName) {
		t.Fatalf("table[%s] should removed now", c.tableName)
	}
}
