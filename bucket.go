package timewheel

import "sync"

type bucket[T any] struct {
	head uint32
	tail uint32
	mu   sync.Mutex
}

func (b *bucket[T]) add(arena *arena[T], nodeIdx uint32) {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := arena.getNode(nodeIdx)

	if b.tail == nullIndex {
		b.head = nodeIdx
		b.tail = nodeIdx
	} else {
		arena.getNode(b.tail).next = nodeIdx
		n.prev = b.tail
		b.tail = nodeIdx
	}
}

func (b *bucket[T]) remove(arena *arena[T], nodeIdx uint32) bool {
	b.mu.Lock()
	defer b.mu.Unlock()

	n := arena.getNode(nodeIdx)

	if n.prev == nullIndex && n.next == nullIndex && b.head != nodeIdx {
		return false
	}

	if n.prev != nullIndex {
		arena.getNode(n.prev).next = n.next
	} else if b.head == nodeIdx {
		b.head = n.next
	}

	if n.next != nullIndex {
		arena.getNode(n.next).prev = n.prev
	} else if b.tail == nodeIdx {
		b.tail = n.prev
	}

	n.next = nullIndex
	n.prev = nullIndex

	return true
}

func (b *bucket[T]) flush() uint32 {
	b.mu.Lock()
	defer b.mu.Unlock()

	head := b.head
	b.head = nullIndex
	b.tail = nullIndex

	return head
}
