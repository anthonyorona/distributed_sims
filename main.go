package main

import (
	"context"
	"math/rand/v2"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Simulate a distributed system, consisting of distinct processes
// that only communicate with one another by exchanging messages.
func main() {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	numProcesses := 5
	processes := make([]*Process, numProcesses)
	randomProcess := rand.IntN(numProcesses) + 1
	initialEvent := Message{
		ProcessID:   ProcessID(randomProcess),
		Timestamp:   Timestamp(0), // must be less than initial clock values
		MessageType: RequestEvent,
	}

	Broadcaster := Broadcaster{
		Queue: make([]Event, 0),
	}
	for i, _ := range processes {
		eventCopy := initialEvent
		var resourceHold *Message
		if i+1 == randomProcess {
			resourceHold = &initialEvent
		}
		processes[i] = &Process{
			ID:           ProcessID(i + 1),
			Broadcaster:  &Broadcaster,
			Directory:    processes,
			RecvChan:     make(chan Event, numProcesses+1),
			EventQueue:   make([]Event, 0),
			Ctx:          ctx,
			Clock:        Timestamp(1),
			DoneWg:       &wg,
			ResourceHold: resourceHold,
		}
		processes[i].RecvChan <- &eventCopy
	}
	go Broadcaster.Start(ctx, processes, &wg)
	for _, p := range processes {
		p.DoneWg.Add(1)
		p.PRoutine()
	}
	<-signalChan // Exit by sending SIG TERM or SIG INT
	cancel()     // Cancel context and wait for clean up
	wg.Wait()
}
