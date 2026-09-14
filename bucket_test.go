package timewheel_test

import (
	"testing"
)

func TestTask_NewWithInt(t *testing.T) {
	//inputTime := time.Now().Add(10 * time.Minute)
	//inputValue := 42
	//
	//task := timewheel.NewTask(inputTime, inputValue)
	//
	//require.Equal(t, inputValue, task.Value)
	//require.Equal(t, inputTime.UnixNano(), task.Expiration)
}

func TestBucket_Sequential(t *testing.T) {
	//bucket := timewheel.NewBucket[string]()
	//
	//require.Nil(t, bucket.Flush())
	//
	//task1 := timewheel.NewTask(time.Now().Add(1*time.Hour), "first-task")
	//task2 := timewheel.NewTask(time.Now().Add(1*time.Hour), "second-task")
	//
	//node1 := bucket.Add(task1)
	//node2 := bucket.Add(task2)
	//
	//evicted := bucket.Flush()
	//
	//require.Len(t, evicted, 2)
	//require.Equal(t, []*timewheel.Node[string]{node1, node2}, evicted)
	//require.Nil(t, bucket.Flush())
}
