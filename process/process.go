package process

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/anthonyorona/logical_clock_sim/common"
	"github.com/anthonyorona/logical_clock_sim/events"
	"github.com/anthonyorona/logical_clock_sim/types"
)

type DirectoryEntry struct {
	ProcessID types.ProcessID
	RecvChan  chan events.RankedEvent
}

type Process struct {
	types.PState
	ProcessWatch
	ID             types.ProcessID
	Ctx            context.Context
	Directory      []DirectoryEntry
	RecvChan       chan events.RankedEvent
	EventSequencer events.EventSequencer
	EventQueue     events.EventQueue[events.RankedEvent]
	rng            *rand.Rand
}

func (p *Process) Broadcast(event events.RankedEvent) {
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
		// pid := types.ProcessID(i)
		// if state, ok := states[pid]; ok {
		// 	fmt.Printf("  PID: %d, LClock: %d, State: %s\n", state.ProcessID, state.Clock, state.State)
		// } else {
		// 	fmt.Printf("  PID: %d, LClock: N/A, State: Unknown (Not yet reported)\n", pid)
		// }
	}
	fmt.Println("----------------------------")
}
