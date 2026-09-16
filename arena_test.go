package timewheel

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestArena_AllocAndFree(t *testing.T) {
	arn := newArena[string](1)

	idx1 := arn.alloc(100, "item1")
	require.NotEqual(t, nullIndex, idx1, "expected valid index, got NullIndex")

	node1 := arn.getNode(idx1)
	assert.Equal(t, int64(100), node1.expiration)
	assert.Equal(t, "item1", node1.value)

	arn.free(idx1)

	assert.Empty(t, node1.value, "expected node value to be zeroed on free")
	assert.Zero(t, node1.expiration, "expected node expiration to be zeroed on free")

	idx2 := arn.alloc(200, "item2")
	assert.Equal(t, idx1, idx2, "expected free list index to be re-used")
}

func TestArena_FreeNullIndex(t *testing.T) {
	arn := newArena[int](1)

	assert.NotPanics(t, func() {
		arn.free(nullIndex)
	})
}

func TestArena_GrowOnDemand(t *testing.T) {
	arn := newArena[int](1)
	initialChunkCount := len(arn.chunks)

	for i := range chunkSize {
		arn.alloc(int64(i), int(i))
	}

	require.Len(t, arn.chunks, initialChunkCount)

	extraIdx := arn.alloc(999, 999)

	assert.Len(t, arn.chunks, initialChunkCount+1)
	assert.Equal(t, chunkSize, extraIdx, "expected first index of next chunk to match ChunkSize")
}

func TestArena_Shrink(t *testing.T) {
	arn := newArena[int](2)

	indices := make([]uint32, chunkSize+10)
	for i := range indices {
		indices[i] = arn.alloc(int64(i), i)
	}

	require.GreaterOrEqual(t, len(arn.chunks), 2)

	for _, idx := range indices {
		if idx >= chunkSize {
			arn.free(idx)
		}
	}

	reclaimed := arn.shrink()
	assert.GreaterOrEqual(t, reclaimed, 1)
	assert.Nil(t, arn.chunks[1], "expected chunk 1 to be nil after Shrink")

	newIdx := arn.alloc(555, 555)
	assert.Equal(t, 555, arn.getNode(newIdx).value)
}

func TestArena_ConcurrentAllocFree(t *testing.T) {
	arn := newArena[int](2)
	const goroutines = 10
	const opsPerGoroutine = 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := range goroutines {
		go func(id int) {
			defer wg.Done()
			allocated := make([]uint32, 0, opsPerGoroutine)

			for i := range opsPerGoroutine {
				idx := arn.alloc(int64(i), id*10000+i)
				allocated = append(allocated, idx)
			}

			for i := 0; i < len(allocated); i += 2 {
				arn.free(allocated[i])
			}
		}(g)
	}

	wg.Wait()
}
