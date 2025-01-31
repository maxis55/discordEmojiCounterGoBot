package utils

func ChunkSlice[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

func ChunkMapValuesByLen[T comparable](input map[T]string, maxChunkLen int) [][]string {
	var chunks [][]string
	var currentChunk []string
	currentSize := 0

	for _, value := range input {
		valueLen := len(value)

		if currentSize+valueLen > maxChunkLen {
			chunks = append(chunks, currentChunk)
			currentChunk = []string{}
			currentSize = 0
		}

		currentChunk = append(currentChunk, value)
		currentSize += valueLen
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

func ChunkSliceValuesByLen(input []string, maxChunkLen int) [][]string {
	var chunks [][]string
	var currentChunk []string
	currentSize := 0

	for _, value := range input {
		valueLen := len(value)

		if currentSize+valueLen > maxChunkLen {
			chunks = append(chunks, currentChunk)
			currentChunk = []string{}
			currentSize = 0
		}

		currentChunk = append(currentChunk, value)
		currentSize += valueLen
	}

	if len(currentChunk) > 0 {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}
