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

	"github.com/anthonyorona/distributed_sims/common"
	"github.com/anthonyorona/distributed_sims/lamport_mutex"
	"github.com/anthonyorona/distributed_sims/process"
	"github.com/anthonyorona/distributed_sims/types"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	var wg sync.WaitGroup
	processWatch := process.ProcessWatch{
		Wg:        &wg,
		PubChange: make(chan process.WatchMessage, 5),
	}

	numProcesses := 10
	processes := make([]*lamport_mutex.LamportProcess, numProcesses)
	randomProcess := rand.IntN(int(numProcesses))
	initialMessage := lamport_mutex.Message{
		ProcessID:   types.ProcessID(randomProcess),
		LTime:       lamport_mutex.LamportTimeStamp(0), // must be less than initial clock values
		MessageType: lamport_mutex.Request,
	}

	var directory []process.DirectoryEntry[*lamport_mutex.Message]
	for i := 0; i < int(numProcesses); i++ {
		pid := types.ProcessID(i)
		recvChan := make(chan *lamport_mutex.Message, numProcesses+1)
		directory = append(directory, process.DirectoryEntry[*lamport_mutex.Message]{
			ProcessID: pid,
			RecvChan:  recvChan,
		})
		currentProcess := lamport_mutex.NewProcess(ctx, pid, initialMessage, recvChan, processWatch)
		if i == randomProcess {
			mCopy := initialMessage
			currentProcess.ResourceHold = &mCopy
			currentProcess.LamportSim.Usage = time.NewTimer(common.GetRandomDuration(1000, 250))
			currentProcess.PState = lamport_mutex.Holding
		} else {
			currentProcess.ResourceHold = nil
			currentProcess.LamportSim.Usage = nil
		}
		processes[i] = &currentProcess
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
		processStates := make(map[types.ProcessID]process.WatchMessage)

		for {
			select {
			case n, ok := <-processWatch.PubChange:
				if !ok {
					return
				}
				processStates[n.ProcessID] = n
				fmt.Printf("(Process, L Clock, State): %d, %s, %s\n", n.ProcessID, n.Clock, n.State)
				process.PrintAllProcessStates(int(numProcesses), processStates)
			case <-ctx.Done():
				return
			}
		}
	}()
	<-ctx.Done()
	processWatch.Wg.Wait()
}
