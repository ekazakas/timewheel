package timewheel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBucket_AddSingleNode(t *testing.T) {
	a := newArena[string](1)
	b := &bucket[string]{
		head: NullIndex,
		tail: NullIndex,
	}

	idx := a.alloc(100, "task-1")
	b.add(a, idx)

	assert.Equal(t, idx, b.head, "head should point to the added node")
	assert.Equal(t, idx, b.tail, "tail should point to the added node")

	nd := a.getNode(idx)
	assert.Equal(t, NullIndex, nd.prev, "single node prev should be nullIndex")
	assert.Equal(t, NullIndex, nd.next, "single node next should be nullIndex")
}

func TestBucket_AddMultipleNodes(t *testing.T) {
	a := newArena[string](1)
	b := &bucket[string]{
		head: NullIndex,
		tail: NullIndex,
	}

	idx1 := a.alloc(100, "task-1")
	idx2 := a.alloc(100, "task-2")
	idx3 := a.alloc(100, "task-3")

	b.add(a, idx1)
	b.add(a, idx2)
	b.add(a, idx3)

	assert.Equal(t, idx1, b.head, "head should be first inserted node")
	assert.Equal(t, idx3, b.tail, "tail should be last inserted node")

	assert.Equal(t, idx2, a.getNode(idx1).next)
	assert.Equal(t, idx3, a.getNode(idx2).next)
	assert.Equal(t, NullIndex, a.getNode(idx3).next)

	assert.Equal(t, NullIndex, a.getNode(idx1).prev)
	assert.Equal(t, idx1, a.getNode(idx2).prev)
	assert.Equal(t, idx2, a.getNode(idx3).prev)
}

func TestBucket_RemoveHeadTailAndMiddle(t *testing.T) {
	a := newArena[int](1)

	t.Run("remove head", func(t *testing.T) {
		b := &bucket[int]{
			head: NullIndex,
			tail: NullIndex,
		}
		idx1 := a.alloc(100, 1)
		idx2 := a.alloc(100, 2)

		b.add(a, idx1)
		b.add(a, idx2)

		removed := b.remove(a, idx1)
		assert.True(t, removed)
		assert.Equal(t, idx2, b.head, "head should update to second node")
		assert.Equal(t, idx2, b.tail)
		assert.Equal(t, NullIndex, a.getNode(idx2).prev)
	})

	t.Run("remove tail", func(t *testing.T) {
		b := &bucket[int]{
			head: NullIndex,
			tail: NullIndex,
		}
		idx1 := a.alloc(100, 1)
		idx2 := a.alloc(100, 2)

		b.add(a, idx1)
		b.add(a, idx2)

		removed := b.remove(a, idx2)
		assert.True(t, removed)
		assert.Equal(t, idx1, b.head)
		assert.Equal(t, idx1, b.tail, "tail should update to first node")
		assert.Equal(t, NullIndex, a.getNode(idx1).next)
	})

	t.Run("remove middle node", func(t *testing.T) {
		b := &bucket[int]{
			head: NullIndex,
			tail: NullIndex,
		}
		idx1 := a.alloc(100, 1)
		idx2 := a.alloc(100, 2)
		idx3 := a.alloc(100, 3)

		b.add(a, idx1)
		b.add(a, idx2)
		b.add(a, idx3)

		removed := b.remove(a, idx2)
		assert.True(t, removed)
		assert.Equal(t, idx1, b.head)
		assert.Equal(t, idx3, b.tail)
		assert.Equal(t, idx3, a.getNode(idx1).next)
		assert.Equal(t, idx1, a.getNode(idx3).prev)
	})

	t.Run("remove non-existent node", func(t *testing.T) {
		b := &bucket[int]{
			head: NullIndex,
			tail: NullIndex,
		}
		idx1 := a.alloc(100, 1)
		idx2 := a.alloc(100, 2)

		b.add(a, idx1)

		removed := b.remove(a, idx2)
		assert.False(t, removed, "removing node not in bucket should return false")
	})
}

func TestBucket_Flush(t *testing.T) {
	a := newArena[string](1)
	b := &bucket[string]{
		head: NullIndex,
		tail: NullIndex,
	}

	idx1 := a.alloc(100, "task-1")
	idx2 := a.alloc(100, "task-2")

	b.add(a, idx1)
	b.add(a, idx2)

	headIdx := b.flush()

	assert.Equal(t, idx1, headIdx, "flush should return original head index")
	assert.Equal(t, NullIndex, b.head, "bucket head should reset to nullIndex")
	assert.Equal(t, NullIndex, b.tail, "bucket tail should reset to nullIndex")

	curr := headIdx
	count := 0
	for curr != NullIndex {
		count++
		curr = a.getNode(curr).next
	}
	assert.Equal(t, 2, count, "traversed linked list count after flush should match added items")
}
