package timewheel

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTimingWheel_AddAndExpire(t *testing.T) {
	start := time.Unix(1000, 0)
	tick := 100 * time.Millisecond
	size := int64(10)

	tw := New[string](tick, size, start)

	var (
		expiredTasks []string
		mu           sync.Mutex
	)
	onExpire := func(val string) {
		mu.Lock()
		expiredTasks = append(expiredTasks, val)
		mu.Unlock()
	}

	exp1 := start.Add(200 * time.Millisecond).UnixNano()
	exp2 := start.Add(500 * time.Millisecond).UnixNano()

	id1 := tw.Add(exp1, "task-200ms")
	id2 := tw.Add(exp2, "task-500ms")

	require.NotEqual(t, nullIndex, id1)
	require.NotEqual(t, nullIndex, id2)

	tw.AdvanceClock(start.Add(300*time.Millisecond).UnixNano(), onExpire)
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, []string{"task-200ms"}, expiredTasks)
	mu.Unlock()

	tw.AdvanceClock(start.Add(600*time.Millisecond).UnixNano(), onExpire)
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	assert.Equal(t, []string{"task-200ms", "task-500ms"}, expiredTasks)
	mu.Unlock()
}

func TestTimingWheel_PastExpiration(t *testing.T) {
	start := time.Unix(1000, 0)
	tw := New[string](time.Second, 10, start)

	pastExp := start.Add(-1 * time.Second).UnixNano()
	idx := tw.Add(pastExp, "past-task")

	assert.Equal(t, nullIndex, idx, "adding task with past expiration should return nullIndex")
}

func TestTimingWheel_Remove(t *testing.T) {
	start := time.Unix(1000, 0)
	tw := New[string](time.Second, 10, start)

	exp := start.Add(3 * time.Second).UnixNano()
	nodeIdx := tw.Add(exp, "cancel-me")
	require.NotEqual(t, nullIndex, nodeIdx)

	removed := tw.Remove(nodeIdx)
	assert.True(t, removed, "remove should return true for active task")

	assert.False(t, tw.Remove(nodeIdx))

	var expiredCount atomic.Int32
	tw.AdvanceClock(start.Add(5*time.Second).UnixNano(), func(val string) {
		expiredCount.Add(1)
	})
	time.Sleep(50 * time.Millisecond)

	assert.Equal(t, int32(0), expiredCount.Load())
}

func TestTimingWheel_OverflowWheelCascade(t *testing.T) {
	start := time.Unix(1000, 0)
	tick := time.Second
	size := int64(10)

	tw := New[string](tick, size, start)

	farExp := start.Add(15 * time.Second).UnixNano()
	nodeIdx := tw.Add(farExp, "overflow-task")
	require.NotEqual(t, nullIndex, nodeIdx)
	require.NotNil(t, tw.overflow, "overflow wheel should be instantiated")

	var (
		fired bool
		wg    sync.WaitGroup
	)
	wg.Add(1)

	onExpire := func(val string) {
		if val == "overflow-task" {
			fired = true
			wg.Done()
		}
	}

	tw.AdvanceClock(start.Add(16*time.Second).UnixNano(), onExpire)
	wg.Wait()

	assert.True(t, fired, "overflow task should fire after cascading down")
}

func TestTimingWheel_ConcurrentAccess(t *testing.T) {
	start := time.Unix(1000, 0)
	tw := New[int](10*time.Millisecond, 100, start)

	const goroutines = 10
	const opsPerGoroutine = 500

	var (
		wg           sync.WaitGroup
		expiredCount atomic.Int64
	)
	wg.Add(goroutines)

	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			for i := range opsPerGoroutine {
				exp := start.Add(time.Duration(i+1) * 20 * time.Millisecond).UnixNano()
				idx := tw.Add(exp, id*1000+i)

				if i%3 == 0 {
					tw.Remove(idx)
				}
			}
		}(g)
	}

	go func() {
		for i := 1; i <= 50; i++ {
			tw.AdvanceClock(start.Add(time.Duration(i*100)*time.Millisecond).UnixNano(), func(val int) {
				expiredCount.Add(1)
			})
			time.Sleep(2 * time.Millisecond)
		}
	}()

	wg.Wait()

	tw.AdvanceClock(start.Add(10*time.Second).UnixNano(), func(val int) {
		expiredCount.Add(1)
	})
	time.Sleep(50 * time.Millisecond)

	assert.Greater(t, expiredCount.Load(), int64(0))
}
