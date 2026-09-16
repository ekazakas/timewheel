package timewheel

import (
	"sync"
	"time"
)

type TimingWheel[T any] struct {
	tick        int64
	size        int64
	interval    int64
	currentTime int64
	buckets     []bucket[T]
	overflow    *TimingWheel[T]
	arena       *arena[T]
	mu          sync.Mutex
}

func New[T any](tick time.Duration, size int64, start time.Time) *TimingWheel[T] {
	return newWheelWithArena[T](tick, size, start, newArena[T](1))
}

func newWheelWithArena[T any](tick time.Duration, size int64, start time.Time, a *arena[T]) *TimingWheel[T] {
	tickNs := tick.Nanoseconds()
	startNs := start.UnixNano()

	buckets := make([]bucket[T], size)
	for i := range size {
		buckets[i] = bucket[T]{
			head: nullIndex,
			tail: nullIndex,
		}
	}

	return &TimingWheel[T]{
		tick:        tickNs,
		size:        size,
		interval:    tickNs * size,
		currentTime: startNs - (startNs % tickNs),
		buckets:     buckets,
		arena:       a,
	}
}

func (tw *TimingWheel[T]) Add(expiration int64, value T) uint32 {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	return tw.add(expiration, value)
}

func (tw *TimingWheel[T]) Remove(nodeIdx uint32) bool {
	if nodeIdx == nullIndex {
		return false
	}

	tw.mu.Lock()
	defer tw.mu.Unlock()

	return tw.remove(nodeIdx)
}

func (tw *TimingWheel[T]) AdvanceClock(targetTimeNs int64, onExpire func(value T)) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	tw.advanceClock(targetTimeNs, onExpire)
}

func (tw *TimingWheel[T]) Shrink() int {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	return tw.arena.shrink()
}

func (tw *TimingWheel[T]) add(expiration int64, value T) uint32 {
	if expiration < tw.currentTime+tw.tick {
		return nullIndex
	}

	if expiration < tw.currentTime+tw.interval {
		idx := (expiration / tw.tick) % tw.size
		nodeIdx := tw.arena.alloc(expiration, value)
		tw.buckets[idx].add(tw.arena, nodeIdx)

		return nodeIdx
	}

	if tw.overflow == nil {
		tw.overflow = newWheelWithArena[T](time.Duration(tw.interval), tw.size, time.Unix(0, tw.currentTime), tw.arena)
	}

	return tw.overflow.add(expiration, value)
}

func (tw *TimingWheel[T]) remove(nodeIdx uint32) bool {
	n := tw.arena.getNode(nodeIdx)

	if n.expiration < tw.currentTime+tw.interval {
		idx := (n.expiration / tw.tick) % tw.size
		removed := tw.buckets[idx].remove(tw.arena, nodeIdx)
		if removed {
			tw.arena.free(nodeIdx)
		}

		return removed
	}

	if tw.overflow != nil {
		return tw.overflow.remove(nodeIdx)
	}

	return false
}

func (tw *TimingWheel[T]) advanceClock(targetTimeNs int64, onExpire func(value T)) {
	for targetTimeNs >= tw.currentTime+tw.tick {
		tw.currentTime += tw.tick

		idx := (tw.currentTime / tw.tick) % tw.size
		head := tw.buckets[idx].flush()

		curr := head
		for curr != nullIndex {
			n := tw.arena.getNode(curr)
			next := n.next
			val := n.value
			exp := n.expiration

			tw.arena.free(curr)

			if exp <= targetTimeNs {
				if onExpire != nil {
					go onExpire(val)
				}
			} else {
				tw.add(exp, val)
			}

			curr = next
		}

		if tw.overflow != nil {
			tw.overflow.advanceClock(tw.currentTime, onExpire)
		}
	}
}
