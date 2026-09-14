package timewheel_test

import (
	"testing"
	"time"

	"github.com/ekazakas/timewheel"
	"github.com/stretchr/testify/require"
)

func TestTickerClock_InterfaceImplementation(t *testing.T) {
	var _ timewheel.Clock = (*timewheel.TickerClock)(nil)

	clock := timewheel.NewTickerClock(1 * time.Millisecond)
	defer clock.Stop()

	require.NotNil(t, clock.TickChan())
	require.WithinDuration(t, time.Now(), clock.Now(), 1*time.Second)
}
