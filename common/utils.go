package common

import (
	"math/rand/v2"
	"time"
)

func GetRandomDuration(baseMillis, varianceMillis int) time.Duration {
	if varianceMillis < 0 {
		varianceMillis = 0
	}
	randomRange := 2*varianceMillis + 1
	randomMillis := rand.IntN(randomRange) - varianceMillis
	totalMillis := baseMillis + randomMillis
	return time.Duration(totalMillis) * time.Millisecond
}
