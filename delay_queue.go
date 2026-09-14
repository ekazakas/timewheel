package timewheel

import (
	"container/heap"
	"sync"
	"time"
)

type (
	DelayQueue[T any] struct {
		pq     bucketHeap[T]
		mu     sync.Mutex
		wakeup chan struct{}
	}

	bucketItem[T any] struct {
		bucket     *Bucket[T]
		expiration int64
		index      int
	}

	bucketHeap[T any] []*bucketItem[T]
)

func (h *bucketHeap[T]) Len() int {
	return len(*h)
}

func (h *bucketHeap[T]) Less(i, j int) bool {
	return (*h)[i].expiration < (*h)[j].expiration
}

func (h *bucketHeap[T]) Swap(i, j int) {
	(*h)[i], (*h)[j] = (*h)[j], (*h)[i]
	(*h)[i].index = i
	(*h)[j].index = j
}

func (h *bucketHeap[T]) Push(x any) {
	item := x.(*bucketItem[T])
	item.index = len(*h)
	*h = append(*h, item)
}

func (h *bucketHeap[T]) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]

	return item
}

func NewDelayQueue[T any](size int) *DelayQueue[T] {
	dq := &DelayQueue[T]{
		pq:     make(bucketHeap[T], 0, size),
		wakeup: make(chan struct{}, 1),
	}

	heap.Init(&dq.pq)

	return dq
}

func (dq *DelayQueue[T]) Offer(b *Bucket[T], expiration int64) {
	dq.mu.Lock()
	defer dq.mu.Unlock()

	item := &bucketItem[T]{
		bucket:     b,
		expiration: expiration,
	}

	heap.Push(&dq.pq, item)

	if item.index == 0 {
		select {
		case dq.wakeup <- struct{}{}:
		default:
		}
	}
}

func (dq *DelayQueue[T]) Poll(exit <-chan struct{}, nowFn func() int64) (*Bucket[T], bool) {
	for {
		dq.mu.Lock()
		if len(dq.pq) == 0 {
			dq.mu.Unlock()
			select {
			case <-exit:
				return nil, false
			case <-dq.wakeup:
				continue
			}
		}

		top := dq.pq[0]
		now := nowFn()
		delta := top.expiration - now

		if delta <= 0 {
			item := heap.Pop(&dq.pq).(*bucketItem[T])
			dq.mu.Unlock()
			return item.bucket, true
		}

		dq.mu.Unlock()

		timer := time.NewTimer(time.Duration(delta))
		select {
		case <-exit:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return nil, false
		case <-dq.wakeup:
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
		case <-timer.C:
		}
	}
}
