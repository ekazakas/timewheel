package timewheel

//
//import (
//	"sync"
//	"time"
//)
//
//type TimingWheel[T any] struct {
//	tick        int64
//	size        int64
//	interval    int64
//	currentTime int64
//	buckets     []Bucket[T]
//	overflow    *TimingWheel[T]
//	dq          *DelayQueue[T]
//	mu          sync.Mutex
//}
//
//func NewTimingWheel[T any](tick time.Duration, start time.Time, size int64, dq *DelayQueue[T]) *TimingWheel[T] {
//	tickNs, startNs := tick.Nanoseconds(), start.UnixNano()
//
//	buckets := make([]Bucket[T], size)
//	//for i := range buckets {
//	//	buckets[i] = NewBucket[T]()
//	//}
//
//	return &TimingWheel[T]{
//		tick:        tickNs,
//		size:        size,
//		interval:    tickNs * size,
//		currentTime: startNs - (startNs % tickNs),
//		buckets:     buckets,
//		dq:          dq,
//	}
//}
//
//func (tw *TimingWheel[T]) Add(task Task[T]) *Node[T] {
//	tw.mu.Lock()
//	defer tw.mu.Unlock()
//
//	return tw.add(task)
//}
//
//func (tw *TimingWheel[T]) Remove(node *Node[T]) bool {
//	if node == nil || node.b == nil {
//		return false
//	}
//
//	tw.mu.Lock()
//	defer tw.mu.Unlock()
//
//	return tw.remove(node)
//}
//
//func (tw *TimingWheel[T]) AdvanceClock(targetTimeNs int64, onExpire func(task Task[T])) {
//	tw.mu.Lock()
//	defer tw.mu.Unlock()
//
//	tw.advanceClock(targetTimeNs, onExpire)
//}
//
//func (tw *TimingWheel[T]) add(task Task[T]) *Node[T] {
//	//if task.Expiration < tw.currentTime+tw.tick {
//	//	return nil
//	//}
//	//
//	//if task.Expiration < tw.currentTime+tw.interval {
//	//	idx := (task.Expiration / tw.tick) % tw.size
//	//	b := tw.buckets[idx]
//	//	node := b.Add(task)
//	//
//	//	bucketExp := task.Expiration - (task.Expiration % tw.tick)
//	//	if b.SetExpiration(bucketExp) {
//	//		tw.dq.Offer(b, bucketExp)
//	//	}
//	//	return node
//	//}
//	//
//	//if tw.overflow == nil {
//	//	tw.overflow = NewTimingWheel[T](time.Duration(tw.interval), time.Unix(0, tw.currentTime), tw.size, tw.dq)
//	//}
//
//	return tw.overflow.add(task)
//}
//
//func (tw *TimingWheel[T]) remove(node *Node[T]) bool {
//	if node == nil || node.b == nil {
//		return false
//	}
//
//	if node.Task.Expiration < tw.currentTime+tw.interval {
//		idx := (node.Task.Expiration / tw.tick) % tw.size
//		return tw.buckets[idx].Remove(node)
//	}
//
//	if tw.overflow != nil {
//		return tw.overflow.remove(node)
//	}
//
//	return false
//}
//
//func (tw *TimingWheel[T]) advanceClock(targetTimeNs int64, onExpire func(task Task[T])) {
//	for targetTimeNs >= tw.currentTime+tw.tick {
//		tw.currentTime += tw.tick
//
//		if tw.overflow != nil {
//			tw.overflow.advanceClock(tw.currentTime, onExpire)
//		}
//	}
//}
