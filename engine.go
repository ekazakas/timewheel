package timewheel

import (
	"context"
	"errors"
	"fmt"
	"time"
)

type Engine[T any] struct {
	wheel *TimingWheel[T]
	clock Clock
	out   chan<- T
}

func NewEngine[T any](wheel *TimingWheel[T], clock Clock, out chan<- T) *Engine[T] {
	return &Engine[T]{
		wheel: wheel,
		clock: clock,
		out:   out,
	}
}

func (e *Engine[T]) Run(ctx context.Context) error {
	tickChan := e.clock.TickChan()
	defer e.clock.Stop()

	for {
		select {
		case <-ctx.Done():
			if err := e.advance(ctx); err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("engine stopped: %w", err)
			}

			return nil
		case _, ok := <-tickChan:
			if !ok {
				return nil
			}

			if err := e.advance(ctx); err != nil {
				if errors.Is(err, context.Canceled) {
					return nil
				}

				return err
			}
		}
	}
}

func (e *Engine[T]) Add(expiration int64, value T) uint32 {
	return e.wheel.Add(expiration, value)
}

func (e *Engine[T]) Schedule(delay time.Duration, value T) uint32 {
	return e.wheel.Add(e.clock.Now().Add(delay).UnixNano(), value)
}

func (e *Engine[T]) Remove(handle uint32) bool {
	return e.wheel.Remove(handle)
}

func (e *Engine[T]) Shrink() int {
	return e.wheel.Shrink()
}

func (e *Engine[T]) advance(ctx context.Context) error {
	due := e.wheel.AdvanceClock(e.clock.Now())

	for _, d := range due {
		select {
		case e.out <- d:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return nil
}
