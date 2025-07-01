package process

import (
	"context"
	"fmt"
	"time"

	"github.com/anthonyorona/distributed_sims/common"
	"github.com/anthonyorona/distributed_sims/events"
	"github.com/anthonyorona/distributed_sims/types"
)

type DirectoryEntry[T events.RankedEvent] struct {
	ProcessID types.ProcessID
	RecvChan  chan T
}

type Process[T events.RankedEvent] struct {
	types.PState
	ProcessWatch
	ID             types.ProcessID
	Ctx            context.Context
	Directory      []DirectoryEntry[T]
	RecvChan       chan T
	EventQueue     events.EventQueue[T]
	EventSequencer events.EventSequencer
}

func (p *Process[T]) Broadcast(event T) {
	time.Sleep(common.GetRandomDuration(200, 100))
	for _, d := range p.Directory {
		if d.ProcessID != event.GetPID() {
			d.RecvChan <- event
		}
	}
}

func PrintAllProcessStates(numProcesses int, states map[types.ProcessID]WatchMessage) {
	fmt.Println("\n--- Current System State ---")
	for i := 0; i < numProcesses; i++ {
		pid := types.ProcessID(i)
		if state, ok := states[pid]; ok {
			fmt.Printf("PID: %d, Clock: %s, Queue Length: %d, State: %s\n", state.ProcessID, state.Clock, state.QL, state.State)
		} else {
			fmt.Printf("  PID: %d, Clock: N/A, State: Unknown (Not yet reported)\n", pid)
		}
	}
	fmt.Println("----------------------------")
}
