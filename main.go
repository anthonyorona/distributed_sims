package main

import (
	"container/heap"
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

	numProcesses := 50
	processes := make([]*Process, numProcesses)
	randomProcess := rand.IntN(numProcesses)
	initialMessage := Message{
		ProcessID:   ProcessID(randomProcess),
		LTime:       LTime(0), // must be less than initial clock values
		MessageType: Request,
	}

	var directory []DirectoryEntry
	for i, _ := range processes {
		pid := ProcessID(i)
		recvChan := make(chan Message, numProcesses+1)
		directory = append(directory, DirectoryEntry{
			ProcessID: pid,
			RecvChan:  recvChan,
		})
		mQueue := make(EventQueue[*Message], 0)
		heap.Init(&mQueue)
		mCopy := initialMessage
		heap.Push(&mQueue, &mCopy)
		currentProcess := &Process{
			ID:              pid,
			RecvChan:        recvChan,
			MessageDispatch: MessageDispatch{},
			MQueue:          mQueue,
			Ctx:             ctx,
			Clock:           LTime(1),
			Sim:             Sim{},
			AckSet:          make(map[ProcessID]struct{}),
			ProcessWatch:    processWatch,
			SRState:         Free,
			rng:             rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), rand.Uint64())),
		}
		if i == randomProcess {
			currentProcess.ResourceHold = &mCopy
			currentProcess.Sim.Usage = time.NewTimer(getRandomDuration(1000, 250))
			currentProcess.SRState = Holding
		} else {
			currentProcess.ResourceHold = nil
			currentProcess.Sim.Usage = nil
		}
		processes[i] = currentProcess
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
		processStates := make(map[ProcessID]WatchMessage)

		for {
			select {
			case n, ok := <-processWatch.PubChange:
				if !ok {
					return
				}
				processStates[n.ProcessID] = n
				fmt.Printf("(Process, L Clock, State): %d, %d, %s\n", n.ProcessID, n.Clock, n.State)
				printAllProcessStates(numProcesses, processStates)
			case <-ctx.Done():
				return
			}
		}
	}()
	<-ctx.Done()
	processWatch.Wg.Wait()
}
