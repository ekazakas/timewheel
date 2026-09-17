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

	exp1 := start.Add(200 * time.Millisecond).UnixNano()
	exp2 := start.Add(500 * time.Millisecond).UnixNano()

	id1 := tw.Add(exp1, "task-200ms")
	id2 := tw.Add(exp2, "task-500ms")

	require.NotEqual(t, NullIndex, id1)
	require.NotEqual(t, NullIndex, id2)

	due1 := tw.AdvanceClock(start.Add(300 * time.Millisecond))
	assert.Equal(t, []string{"task-200ms"}, due1)

	due2 := tw.AdvanceClock(start.Add(600 * time.Millisecond))
	assert.Equal(t, []string{"task-500ms"}, due2)
}

func TestTimingWheel_PastExpiration(t *testing.T) {
	start := time.Unix(1000, 0)
	tw := New[string](time.Second, 10, start)

	pastExp := start.Add(-1 * time.Second).UnixNano()
	idx := tw.Add(pastExp, "past-task")

	assert.Equal(t, NullIndex, idx, "adding task with past expiration should return nullIndex")
}

func TestTimingWheel_Remove(t *testing.T) {
	start := time.Unix(1000, 0)
	tw := New[string](time.Second, 10, start)

	exp := start.Add(3 * time.Second).UnixNano()
	nodeIdx := tw.Add(exp, "cancel-me")
	require.NotEqual(t, NullIndex, nodeIdx)

	removed := tw.Remove(nodeIdx)
	assert.True(t, removed, "remove should return true for active task")
	assert.False(t, tw.Remove(nodeIdx), "subsequent remove should return false")

	due := tw.AdvanceClock(start.Add(5 * time.Second))
	assert.Empty(t, due, "removed item should not be returned on expiration")
}

func TestTimingWheel_OverflowWheelCascade(t *testing.T) {
	start := time.Unix(1000, 0)
	tick := time.Second
	size := int64(10)

	tw := New[string](tick, size, start)

	farExp := start.Add(15 * time.Second).UnixNano()
	nodeIdx := tw.Add(farExp, "overflow-task")
	require.NotEqual(t, NullIndex, nodeIdx)
	require.NotNil(t, tw.overflow, "overflow wheel should be instantiated")

	due := tw.AdvanceClock(start.Add(16 * time.Second))
	assert.Contains(t, due, "overflow-task", "overflow task should fire after cascading down")
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

	stopTicker := make(chan struct{})
	go func() {
		for i := 1; i <= 50; i++ {
			select {
			case <-stopTicker:
				return
			default:
				due := tw.AdvanceClock(start.Add(time.Duration(i*100) * time.Millisecond))
				expiredCount.Add(int64(len(due)))
				time.Sleep(2 * time.Millisecond)
			}
		}
	}()

	wg.Wait()
	close(stopTicker)

	due := tw.AdvanceClock(start.Add(10 * time.Second))
	expiredCount.Add(int64(len(due)))

	assert.Greater(t, expiredCount.Load(), int64(0))
}
