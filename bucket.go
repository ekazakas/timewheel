package timewheel

type bucket[T any] struct {
	head uint32
	tail uint32
}

func (b *bucket[T]) add(arena *arena[T], nodeIdx uint32) {
	n := arena.getNode(nodeIdx)

	if b.tail == NullIndex {
		b.head = nodeIdx
		b.tail = nodeIdx
	} else {
		arena.getNode(b.tail).next = nodeIdx
		n.prev = b.tail
		b.tail = nodeIdx
	}
}

func (b *bucket[T]) remove(arena *arena[T], nodeIdx uint32) bool {
	n := arena.getNode(nodeIdx)

	if n.prev == NullIndex && n.next == NullIndex && b.head != nodeIdx {
		return false
	}

	if n.prev != NullIndex {
		arena.getNode(n.prev).next = n.next
	} else if b.head == nodeIdx {
		b.head = n.next
	}

	if n.next != NullIndex {
		arena.getNode(n.next).prev = n.prev
	} else if b.tail == nodeIdx {
		b.tail = n.prev
	}

	n.next = NullIndex
	n.prev = NullIndex

	return true
}

func (b *bucket[T]) flush() uint32 {
	head := b.head
	b.head = NullIndex
	b.tail = NullIndex

	return head
}
