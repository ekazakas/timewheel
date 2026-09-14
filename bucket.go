package timewheel

//
//const (
//	NullIndex            = 0
//	DefaultArenaCapacity = 1024
//)
//
//type (
//	Arena[T any] struct {
//		nodes     []Node[T]
//		freeHead  uint32
//		freeCount uint32
//	}
//
//	Node[T any] struct {
//		Expiration int64
//		Value      T
//		Next       uint32
//		Prev       uint32
//	}
//
//	Bucket[T any] struct {
//		Head uint32
//		Tail uint32
//	}
//)
//
//func NewArena[T any](capacity uint32) *Arena[T] {
//	if capacity == 0 {
//		capacity = DefaultArenaCapacity
//	}
//
//	return &Arena[T]{
//		nodes: make([]Node[T], 1, capacity+1),
//	}
//}
//
//func (a *Arena[T]) Alloc(expiration int64, value T) uint32 {
//	if a.freeHead != NullIndex {
//		idx := a.freeHead
//		a.freeHead = a.nodes[idx].Next
//		a.freeCount--
//
//		a.nodes[idx] = Node[T]{
//			Expiration: expiration,
//			Value:      value,
//			Next:       NullIndex,
//			Prev:       NullIndex,
//		}
//
//		return idx
//	}
//
//	idx := uint32(len(a.nodes))
//
//	a.nodes = append(a.nodes, Node[T]{
//		Expiration: expiration,
//		Value:      value,
//		Next:       NullIndex,
//		Prev:       NullIndex,
//	})
//
//	return idx
//}
//
//func (a *Arena[T]) Free(idx uint32) {
//	if idx == NullIndex || idx >= uint32(len(a.nodes)) {
//		return
//	}
//
//	var zero T
//
//	a.nodes[idx] = Node[T]{
//		Value: zero,
//		Next:  a.freeHead,
//	}
//	a.freeHead = idx
//	a.freeCount++
//}
//
//func (b *Bucket[T]) Add(arena *Arena[T], nodeIdx uint32) {
//	if b.Tail == NullIndex {
//		b.Head = nodeIdx
//		b.Tail = nodeIdx
//	} else {
//		arena.nodes[b.Tail].Next = nodeIdx
//		arena.nodes[nodeIdx].Prev = b.Tail
//		b.Tail = nodeIdx
//	}
//}
//
//func (b *Bucket[T]) Remove(arena *Arena[T], nodeIdx uint32) bool {
//	if nodeIdx == NullIndex || nodeIdx >= uint32(len(arena.nodes)) {
//		return false
//	}
//
//	node := &arena.nodes[nodeIdx]
//
//	if node.Prev == NullIndex && node.Next == NullIndex && b.Head != nodeIdx {
//		return false
//	}
//
//	if node.Prev != NullIndex {
//		arena.nodes[node.Prev].Next = node.Next
//	} else if b.Head == nodeIdx {
//		b.Head = node.Next
//	}
//
//	if node.Next != NullIndex {
//		arena.nodes[node.Next].Prev = node.Prev
//	} else if b.Tail == nodeIdx {
//		b.Tail = node.Prev
//	}
//
//	node.Next = NullIndex
//	node.Prev = NullIndex
//
//	return true
//}
//
//func (b *Bucket[T]) Flush() uint32 {
//	head := b.Head
//
//	b.Head = NullIndex
//	b.Tail = NullIndex
//
//	return head
//}
