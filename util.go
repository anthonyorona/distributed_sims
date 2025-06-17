package main

import (
	"math/rand/v2"
	"time"
)

func RemoveAtIndices[T interface{}](s []T, indices map[int]struct{}) []T {
	var updatedSlice []T
	for i, e := range s {
		if _, found := indices[i]; !found {
			updatedSlice = append(updatedSlice, e)
		}
	}
	return updatedSlice
}

func RemoveWhere[T any](s []T, where func(T) bool) []T {
	var result []T
	for _, element := range s {
		if !where(element) {
			result = append(result, element)
		}
	}
	return result
}

func getRandomDuration(baseMillis, varianceMillis int) time.Duration {
	if varianceMillis < 0 {
		varianceMillis = 0
	}
	randomRange := 2*varianceMillis + 1
	randomMillis := rand.IntN(randomRange) - varianceMillis
	totalMillis := baseMillis + randomMillis
	return time.Duration(totalMillis) * time.Millisecond
}
