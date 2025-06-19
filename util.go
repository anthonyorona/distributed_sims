package main

import (
	"fmt"
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

func printAllProcessStates(numProcesses int, states map[ProcessID]WatchMessage) {
	fmt.Println("\n--- Current System State ---")
	for i := 0; i < numProcesses; i++ {
		pid := ProcessID(i)
		if state, ok := states[pid]; ok {
			fmt.Printf("  PID: %d, LClock: %d, State: %s\n", state.ProcessID, state.Clock, state.State)
		} else {
			fmt.Printf("  PID: %d, LClock: N/A, State: Unknown (Not yet reported)\n", pid)
		}
	}
	fmt.Println("----------------------------")
}
