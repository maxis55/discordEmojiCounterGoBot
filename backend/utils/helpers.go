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

func RedistributeSlicesIntoAmountBasedOnEntries(initial [][]string, newSize int) [][]string {
	var allStrings []string
	for _, slice := range initial {
		allStrings = append(allStrings, slice...)
	}

	totalLength := len(allStrings)

	baseSize := totalLength / newSize
	columnWithExtra := totalLength % newSize

	newSlices := make([][]string, newSize)

	index := 0
	for i := 0; i < newSize; i++ {
		limit := baseSize
		if i < columnWithExtra {
			limit++
		}

		newSlices[i] = allStrings[index : index+limit]
		index += limit
	}

	return newSlices
}

func RedistributeSlicesBasedOnMaxLen(initial [][]string, maxLength int) [][]string {
	// skip the last slice to avoid overflow
	for i := 0; i < len(initial)-1; i++ {
		currentLen := 0
		for _, str := range initial[i] {
			currentLen += len(str)
		}

		for currentLen > maxLength {
			lastEntryIndex := len(initial[i]) - 1
			lastEntry := initial[i][lastEntryIndex]
			initial[i] = initial[i][:lastEntryIndex]
			initial[i+1] = append([]string{lastEntry}, initial[i+1]...)

			currentLen -= len(lastEntry)
		}
	}

	return initial
}

func RedistributeSlicesIntoAmountBasedOnLength(initial [][]string, newSize int) [][]string {
	totalLength := 0
	for _, slice := range initial {
		for _, str := range slice {
			totalLength += len(str)
		}
	}
	idealLength := totalLength / newSize

	newSlices := make([][]string, newSize)

	currentSlice := 0
	currentLength := 0
	firstSliceFilled := false
	for _, slice := range initial {
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
