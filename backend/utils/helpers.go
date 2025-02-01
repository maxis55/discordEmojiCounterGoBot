package utils

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

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

func ChunkSlice[T any](items []T, chunkSize int) (chunks [][]T) {
	for chunkSize < len(items) {
		items, chunks = items[chunkSize:], append(chunks, items[0:chunkSize:chunkSize])
	}
	return append(chunks, items)
}

func RedistributeSlicesIntoAmount(chunk [][]string, newSize int) [][]string {
	totalLength := 0
	for _, slice := range chunk {
		for _, str := range slice {
			totalLength += len(str)
		}
	}
	idealLength := totalLength / newSize

	newSlices := make([][]string, newSize)

	currentSlice := 0
	currentLength := 0
	firstSliceFilled := false
	for _, slice := range chunk {
		for _, str := range slice {
			if currentLength+len(str) > idealLength && currentSlice < newSize-1 && firstSliceFilled {
				currentSlice++
				currentLength = 0
			}
			newSlices[currentSlice] = append(newSlices[currentSlice], str)
			currentLength += len(str)
			firstSliceFilled = true
		}
	}

	return newSlices
}

func GetMD5Hash(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}

func SQLPlaceholders(n int) string {
	parts := make([]string, n)
	for i := range parts {
		parts[i] = fmt.Sprintf("$%d", i+1)
	}
	return strings.Join(parts, ", ")
}

func GetKeysFromMap[T comparable](m map[T]any) []T {
	keys := make([]T, 0, len(m))

	for key := range m {
		keys = append(keys, key)
	}

	return keys
}

func GetValuesFromMapBasedOnKeys[T comparable](m map[T]any, keys []T) []any {
	values := make([]any, len(keys))

	for i, key := range keys {
		values[i] = m[key]
	}

	return values
}

func GetEnvStrWithFallback(envKey string, fallback string) string {
	if v := os.Getenv(envKey); v != "" {
		return v
	}

	return fallback
}
