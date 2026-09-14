package timewheel

import (
	"sync"
	"time"
)

const (
	NullIndex            = 0
	DefaultArenaCapacity = 8096
)

type (
	Task[T any] struct {
		Expiration int64 // Nanoseconds
		Value      T
	}

	Node[T any] struct {
		Task Task[T]
		Next uint32
		Prev uint32
	}

	Arena[T any] struct {
		nodes     []Node[T]
		freeHead  uint32
		freeCount uint32
	}

	Bucket[T any] struct {
		Head uint32
		Tail uint32
	}

	TimingWheel[T any] struct {
		tick        int64
		size        int64
		interval    int64
		currentTime int64

		buckets  []Bucket[T]
		overflow *TimingWheel[T]
		arena    *Arena[T]

		mu sync.Mutex
	}
)

func NewArena[T any](capacity uint32) *Arena[T] {
	if capacity == 0 {
		capacity = DefaultArenaCapacity
	}

	return &Arena[T]{
		nodes: make([]Node[T], 1, capacity+1),
	}
}

func NewTimingWheel[T any](tick time.Duration, size int64, start time.Time, arena *Arena[T]) *TimingWheel[T] {
	tickNs := tick.Nanoseconds()
	startNs := start.UnixNano()

	if arena == nil {
		arena = NewArena[T](8096)
	}

	return &TimingWheel[T]{
		tick:        tickNs,
		size:        size,
		interval:    tickNs * size,
		currentTime: startNs - (startNs % tickNs),
		buckets:     make([]Bucket[T], size),
		arena:       arena,
	}
}

func (a *Arena[T]) Alloc(task Task[T]) uint32 {
	if a.freeHead != NullIndex {
		idx := a.freeHead
		a.freeHead = a.nodes[idx].Next
		a.freeCount--

		a.nodes[idx] = Node[T]{
			Task: task,
			Next: NullIndex,
			Prev: NullIndex,
		}

		return idx
	}

	idx := uint32(len(a.nodes))

	a.nodes = append(a.nodes, Node[T]{
		Task: task,
		Next: NullIndex,
		Prev: NullIndex,
	})

	return idx
}

func (a *Arena[T]) Free(idx uint32) {
	if idx == NullIndex || idx >= uint32(len(a.nodes)) {
		return
	}

	var zero T

	a.nodes[idx] = Node[T]{
		Task: Task[T]{
			Value: zero,
		},
		Next: a.freeHead,
	}
	a.freeHead = idx
	a.freeCount++
}

func (b *Bucket[T]) Add(arena *Arena[T], nodeIdx uint32) {
	if b.Tail == NullIndex {
		b.Head = nodeIdx
		b.Tail = nodeIdx
	} else {
		arena.nodes[b.Tail].Next = nodeIdx
		arena.nodes[nodeIdx].Prev = b.Tail
		b.Tail = nodeIdx
	}
}

func (b *Bucket[T]) Remove(arena *Arena[T], nodeIdx uint32) bool {
	if nodeIdx == NullIndex || nodeIdx >= uint32(len(arena.nodes)) {
		return false
	}

	node := &arena.nodes[nodeIdx]

	if node.Prev == NullIndex && node.Next == NullIndex && b.Head != nodeIdx {
		return false
	}

	if node.Prev != NullIndex {
		arena.nodes[node.Prev].Next = node.Next
	} else if b.Head == nodeIdx {
		b.Head = node.Next
	}

	if node.Next != NullIndex {
		arena.nodes[node.Next].Prev = node.Prev
	} else if b.Tail == nodeIdx {
		b.Tail = node.Prev
	}

	node.Next = NullIndex
	node.Prev = NullIndex

	return true
}

func (b *Bucket[T]) Flush() uint32 {
	head := b.Head

	b.Head = NullIndex
	b.Tail = NullIndex

	return head
}

func (tw *TimingWheel[T]) Add(task Task[T]) uint32 {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	return tw.add(task)
}

func (tw *TimingWheel[T]) Remove(nodeIdx uint32) bool {
	if nodeIdx == NullIndex {
		return false
	}

	tw.mu.Lock()
	defer tw.mu.Unlock()

	return tw.remove(nodeIdx)
}

func (tw *TimingWheel[T]) AdvanceClock(targetTimeNs int64, onExpire func(task Task[T])) {
	tw.mu.Lock()
	defer tw.mu.Unlock()

	tw.advanceClock(targetTimeNs, onExpire)
}

func (tw *TimingWheel[T]) add(task Task[T]) uint32 {
	if task.Expiration < tw.currentTime+tw.tick {
		return NullIndex
	}

	if task.Expiration < tw.currentTime+tw.interval {
		idx := (task.Expiration / tw.tick) % tw.size
		nodeIdx := tw.arena.Alloc(task)

		tw.buckets[idx].Add(tw.arena, nodeIdx)

		return nodeIdx
	}

	if tw.overflow == nil {
		tw.overflow = NewTimingWheel[T](
			time.Duration(tw.interval),
			tw.size,
			time.Unix(0, tw.currentTime),
			tw.arena,
		)
	}

	return tw.overflow.add(task)
}

func (tw *TimingWheel[T]) remove(nodeIdx uint32) bool {
	if nodeIdx >= uint32(len(tw.arena.nodes)) {
		return false
	}

	task := tw.arena.nodes[nodeIdx].Task

	if task.Expiration < tw.currentTime+tw.interval {
		idx := (task.Expiration / tw.tick) % tw.size
		removed := tw.buckets[idx].Remove(tw.arena, nodeIdx)
		if removed {
			tw.arena.Free(nodeIdx)
		}
		return removed
	}

	if tw.overflow != nil {
		return tw.overflow.remove(nodeIdx)
	}

	return false
}

func (tw *TimingWheel[T]) advanceClock(targetTimeNs int64, onExpire func(task Task[T])) {
	for targetTimeNs >= tw.currentTime+tw.tick {
		tw.currentTime += tw.tick

		idx := (tw.currentTime / tw.tick) % tw.size
		head := tw.buckets[idx].Flush()

		curr := head
		for curr != NullIndex {
			next := tw.arena.nodes[curr].Next
			task := tw.arena.nodes[curr].Task

			tw.arena.Free(curr)

			if task.Expiration <= targetTimeNs {
				if onExpire != nil {
					onExpire(task)
				}
			} else {
				tw.add(task)
			}

			curr = next
		}

		if tw.overflow != nil {
			tw.overflow.advanceClock(tw.currentTime, onExpire)
		}
	}
}
