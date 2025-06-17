package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Simulate a distributed system, consisting of distinct processes
// that only communicate with one another by exchanging messages.
func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	var wg sync.WaitGroup
	processWatch := ProcessWatch{
		Wg:        &wg,
		PubChange: make(chan WatchMessage, 5),
	}

	numProcesses := 5
	processes := make([]*Process, numProcesses)
	randomProcess := rand.IntN(numProcesses) + 1
	initialMessage := Message{
		ProcessID:   ProcessID(randomProcess),
		LTime:       LTime(0), // must be less than initial clock values
		MessageType: Request,
	}

	var directory []DirectoryEntry
	for i, _ := range processes {
		PID := ProcessID(i)
		var resourceHold Message
		var simulateUsage *time.Timer
		recvChan := make(chan Message, numProcesses+1)
		directory = append(directory, DirectoryEntry{
			ProcessID: PID,
			RecvChan:  recvChan,
		})
		if i+1 == randomProcess {
			mCopy := initialMessage
			processes[i] = &Process{
				ID:           PID,
				RecvChan:     recvChan,
				MQueue:       make([]Message, 0),
				Ctx:          ctx,
				Clock:        LTime(1),
				ResourceHold: &mCopy,
				Sim: Sim{
					Usage: time.NewTimer(getRandomDuration(151, 100)),
				},
				AckSet:       make(map[ProcessID]struct{}),
				ProcessWatch: processWatch,
				SRState:      Holding,
			}
		} else {
			processes[i] = &Process{
				ID:           PID,
				RecvChan:     recvChan,
				MQueue:       make([]Message, 0),
				Ctx:          ctx,
				Clock:        LTime(1),
				ResourceHold: &resourceHold,
				Sim: Sim{
					Usage: simulateUsage,
				},
				AckSet:       make(map[ProcessID]struct{}),
				ProcessWatch: processWatch,
				SRState:      Free,
			}
		}
		processes[i].RecvChan <- initialMessage
	}

	for _, p := range processes {
		processWatch.Wg.Add(1)
		p.Directory = directory
		go p.Start()
	}
	processWatch.Wg.Add(1)
	go func() {
		defer processWatch.Wg.Done()
		defer close(processWatch.PubChange)
		for {
			select {
			case n, ok := <-processWatch.PubChange:
				if !ok {
					fmt.Printf("(Process, L Clock, State): %d, %d, %s", n.ProcessID, n.Clock, n.State)
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	<-ctx.Done()
	processWatch.Wg.Wait()
}
