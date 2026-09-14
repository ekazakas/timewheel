package timewheel

type (
	Node[T any] struct {
		Task Task[T]
		next *Node[T]
		prev *Node[T]
		b    *Bucket[T]
	}

	Bucket[T any] struct {
		head       *Node[T]
		tail       *Node[T]
		expiration int64
	}
)

func NewBucket[T any]() *Bucket[T] {
	return &Bucket[T]{
		expiration: -1,
	}
}

func (b *Bucket[T]) Add(task Task[T]) *Node[T] {
	node := &Node[T]{
		Task: task,
		b:    b,
	}

	if b.tail == nil {
		b.head = node
		b.tail = node
	} else {
		b.tail.next = node
		node.prev = b.tail
		b.tail = node
	}

	return node
}

func (b *Bucket[T]) Remove(node *Node[T]) bool {
	if node == nil || node.b != b {
		return false
	}

	if node.prev != nil {
		node.prev.next = node.next
	} else {
		b.head = node.next
	}

	if node.next != nil {
		node.next.prev = node.prev
	} else {
		b.tail = node.prev
	}

	node.next = nil
	node.prev = nil
	node.b = nil

	return true
}

func (b *Bucket[T]) Flush(fn func(node *Node[T])) {
	curr := b.head
	b.head = nil
	b.tail = nil
	b.expiration = -1

	for curr != nil {
		next := curr.next
		curr.next = nil
		curr.prev = nil
		curr.b = nil

		fn(curr)
		curr = next
	}
}

func (b *Bucket[T]) SetExpiration(exp int64) bool {
	if b.expiration != exp {
		b.expiration = exp

		return true
	}

	return false
}

func (b *Bucket[T]) Expiration() int64 {
	return b.expiration
}
