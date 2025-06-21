package process

import (
	"context"
	"fmt"
	"time"

	"github.com/anthonyorona/logical_clock_sim/common"
	"github.com/anthonyorona/logical_clock_sim/events"
	"github.com/anthonyorona/logical_clock_sim/types"
)

type DirectoryEntry struct {
	ProcessID types.ProcessID
	RecvChan  chan events.RankedEvent
}

type Process[T events.RankedEvent] struct {
	types.PState
	ProcessWatch
	ID             types.ProcessID
	Ctx            context.Context
	Directory      []DirectoryEntry
	RecvChan       chan T
	EventQueue     events.EventQueue[T]
	EventSequencer events.EventSequencer
}

func (p *Process[T]) Broadcast(event events.RankedEvent) {
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
			fmt.Printf("  PID: %d, Clock: %s, State: %s\n", state.ProcessID, state.Clock, state.State)
		} else {
			fmt.Printf("  PID: %d, Clock: N/A, State: Unknown (Not yet reported)\n", pid)
		}
	}
	fmt.Println("----------------------------")
}
