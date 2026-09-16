package timewheel

const (
	nullIndex  uint32 = 0xFFFFFFFF
	chunkShift uint32 = 16
	chunkSize  uint32 = 1 << chunkShift // 65,536 elements per slab
	chunkMask  uint32 = chunkSize - 1
)
