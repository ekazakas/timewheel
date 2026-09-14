package timewheel

import (
	"sync"
	"time"
)

type Scheduler[T any] struct {
	tw   *TimingWheel[T]
	dq   *DelayQueue[T]
	exit chan struct{}
	wg   sync.WaitGroup
}

func NewScheduler[T any](tick time.Duration, wheelSize int64) *Scheduler[T] {
	dq := NewDelayQueue[T](int(wheelSize))
	tw := NewTimingWheel[T](tick, time.Now(), wheelSize, dq)

	return &Scheduler[T]{
		tw:   tw,
		dq:   dq,
		exit: make(chan struct{}),
	}
}

func (s *Scheduler[T]) Add(task Task[T]) *Node[T] {
	return s.tw.Add(task)
}

func (s *Scheduler[T]) Remove(node *Node[T]) bool {
	return s.tw.Remove(node)
}

func (s *Scheduler[T]) Start(onExpire func(task Task[T])) {
	//s.wg.Add(1)
	//go func() {
	//	defer s.wg.Done()
	//	for {
	//		bucket, ok := s.dq.Poll(s.exit, func() int64 { return time.Now().UnixNano() })
	//		if !ok {
	//			return
	//		}
	//
	//		s.tw.AdvanceClock(bucket.Expiration(), onExpire)
	//
	//		bucket.Flush(func(node *Node[T]) {
	//			if s.tw.Add(node.Task) == nil {
	//				onExpire(node.Task)
	//			}
	//		})
	//	}
	//}()
}

func (s *Scheduler[T]) Stop() {
	close(s.exit)
	s.wg.Wait()
}
