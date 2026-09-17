package timewheel

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockClock struct {
	now time.Time
	ch  chan time.Time
	mu  sync.Mutex
}

func NewMockClock(start time.Time) *MockClock {
	return &MockClock{
		now: start,
		ch:  make(chan time.Time, 100),
	}
}

func (m *MockClock) Now() time.Time {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.now
}

func (m *MockClock) TickChan() <-chan time.Time {
	return m.ch
}

func (m *MockClock) Advance(d time.Duration) {
	m.mu.Lock()
	m.now = m.now.Add(d)
	current := m.now
	m.mu.Unlock()

	m.ch <- current
}

func (m *MockClock) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	select {
	case <-m.ch:
	default:
		close(m.ch)
	}
}

func TestEngine_ScheduleAndExpire(t *testing.T) {
	start := time.Unix(1000, 0)
	clock := NewMockClock(start)
	out := make(chan string, 10)

	wheel := NewTimingWheel[string](100*time.Millisecond, 10, start)
	engine := NewEngine[string](wheel, clock, out)

	ctx := t.Context()

	go func() {
		err := engine.Run(ctx)
		assert.NoError(t, err)
	}()

	id1 := engine.Schedule(200*time.Millisecond, "job-200ms")
	id2 := engine.Schedule(500*time.Millisecond, "job-500ms")

	require.NotEqual(t, NullIndex, id1)
	require.NotEqual(t, NullIndex, id2)

	clock.Advance(300 * time.Millisecond)

	select {
	case val := <-out:
		assert.Equal(t, "job-200ms", val)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for job-200ms")
	}

	clock.Advance(300 * time.Millisecond)

	select {
	case val := <-out:
		assert.Equal(t, "job-500ms", val)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for job-500ms")
	}
}

func TestEngine_Cancel(t *testing.T) {
	start := time.Unix(1000, 0)
	clock := NewMockClock(start)
	out := make(chan string, 10)

	wheel := NewTimingWheel[string](100*time.Millisecond, 10, start)
	engine := NewEngine[string](wheel, clock, out)

	ctx := t.Context()

	go func() {
		_ = engine.Run(ctx)
	}()

	id := engine.Schedule(200*time.Millisecond, "canceled-job")
	require.NotEqual(t, NullIndex, id)

	removed := engine.Remove(id)
	assert.True(t, removed)

	clock.Advance(300 * time.Millisecond)

	select {
	case val := <-out:
		t.Fatalf("unexpected value received: %s", val)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestEngine_ContextCancellationUnblocksFullChannel(t *testing.T) {
	start := time.Unix(1000, 0)
	clock := NewMockClock(start)
	out := make(chan string, 1)

	wheel := NewTimingWheel[string](100*time.Millisecond, 10, start)
	engine := NewEngine[string](wheel, clock, out)

	ctx, cancel := context.WithCancel(context.Background())

	engine.Schedule(100*time.Millisecond, "item-1")
	engine.Schedule(100*time.Millisecond, "item-2")

	engineDone := make(chan error, 1)
	go func() {
		engineDone <- engine.Run(ctx)
	}()

	clock.Advance(100 * time.Millisecond)

	time.Sleep(10 * time.Millisecond)
	cancel()

	select {
	case err := <-engineDone:
		assert.NoError(t, err)
	case <-time.After(1 * time.Second):
		t.Fatal("Engine.Run deadlocked when out channel was full during cancellation")
	}
}

func TestEngine_ClockChannelClosed(t *testing.T) {
	start := time.Unix(1000, 0)
	clock := NewMockClock(start)
	out := make(chan string, 10)

	wheel := NewTimingWheel[string](100*time.Millisecond, 10, start)
	engine := NewEngine[string](wheel, clock, out)

	engineDone := make(chan error, 1)
	go func() {
		engineDone <- engine.Run(context.Background())
	}()

	clock.Stop()

	select {
	case err := <-engineDone:
		assert.NoError(t, err, "engine should exit cleanly when clock channel is closed")
	case <-time.After(1 * time.Second):
		t.Fatal("engine failed to exit after clock channel closed")
	}
}
