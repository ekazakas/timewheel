package timewheel_test

import (
	"testing"
	"time"

	"github.com/ekazakas/timewheel"
	"github.com/stretchr/testify/require"
)

func TestBucket_Sequential(t *testing.T) {
	bucket := timewheel.NewBucket[string]()

	require.Nil(t, bucket.Flush())

	task1 := timewheel.NewTask(time.Now().Add(1*time.Hour), "first-task")
	task2 := timewheel.NewTask(time.Now().Add(1*time.Hour), "second-task")

	node1 := bucket.Add(task1)
	node2 := bucket.Add(task2)

	evicted := bucket.Flush()

	require.Len(t, evicted, 2)
	require.Equal(t, []*timewheel.Node[string]{node1, node2}, evicted)
	require.Nil(t, bucket.Flush())
}
