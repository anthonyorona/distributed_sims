package main

import (
	"math/rand/v2"
	"time"
)

func getRandomDuration(baseMillis, varianceMillis int) time.Duration {
	if varianceMillis < 0 {
		varianceMillis = 0
	}
	randomRange := 2*varianceMillis + 1
	randomMillis := rand.IntN(randomRange) - varianceMillis
	totalMillis := baseMillis + randomMillis
	return time.Duration(totalMillis) * time.Millisecond
}
