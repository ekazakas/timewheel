package timewheel

import (
	"sync"
)

type (
	arena[T any] struct {
		chunks      [][]node[T]
		activeCount []uint32
		freeHead    uint32
		top         uint32
		mu          sync.Mutex
	}

	node[T any] struct {
		expiration int64
		value      T
		next       uint32
		prev       uint32
	}
)

func newArena[T any](initialChunks int) *arena[T] {
	if initialChunks <= 0 {
		initialChunks = 1
	}

	a := &arena[T]{
		chunks:      make([][]node[T], 0, initialChunks),
		activeCount: make([]uint32, 0, initialChunks),
		freeHead:    nullIndex,
		top:         0,
	}

	for i := 0; i < initialChunks; i++ {
		a.grow()
	}

	return a
}

func (a *arena[T]) grow() {
	newChunk := make([]node[T], chunkSize)
	a.chunks = append(a.chunks, newChunk)
	a.activeCount = append(a.activeCount, 0)
}

func (a *arena[T]) getNode(idx uint32) *node[T] {
	return &a.chunks[idx>>chunkShift][idx&chunkMask]
}

func (a *arena[T]) alloc(expiration int64, value T) uint32 {
	a.mu.Lock()
	defer a.mu.Unlock()

	var idx uint32

	if a.freeHead != nullIndex {
		idx = a.freeHead
		chunkIdx := idx >> chunkShift

		if a.chunks[chunkIdx] == nil {
			a.chunks[chunkIdx] = make([]node[T], chunkSize)
		}

		n := a.getNode(idx)
		a.freeHead = n.next
		a.activeCount[chunkIdx]++

		n.expiration = expiration
		n.value = value
		n.next = nullIndex
		n.prev = nullIndex

		return idx
	}

	chunkIdx := a.top >> chunkShift

	for chunkIdx >= uint32(len(a.chunks)) {
		a.grow()
	}

	if a.chunks[chunkIdx] == nil {
		a.chunks[chunkIdx] = make([]node[T], chunkSize)
	}

	idx = a.top
	a.top++

	a.activeCount[chunkIdx]++

	n := a.getNode(idx)
	n.expiration = expiration
	n.value = value
	n.next = nullIndex
	n.prev = nullIndex

	return idx
}

func (a *arena[T]) free(idx uint32) {
	if idx == nullIndex {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	chunkIdx := idx >> chunkShift
	n := a.getNode(idx)

	var zero T
	n.value = zero
	n.expiration = 0
	n.prev = nullIndex
	n.next = a.freeHead

	a.freeHead = idx
	a.activeCount[chunkIdx]--
}

func (a *arena[T]) shrink() int {
	a.mu.Lock()
	defer a.mu.Unlock()

	reclaimed := 0

	for i := 1; i < len(a.chunks); i++ {
		if a.activeCount[i] == 0 && a.chunks[i] != nil {
			a.chunks[i] = nil
			reclaimed++
		}
	}

	return reclaimed
}
