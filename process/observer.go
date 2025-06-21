package process

import (
	"sync"

	"github.com/anthonyorona/logical_clock_sim/types"
)

type WatchMessage struct {
	ProcessID types.ProcessID
	State     string
	Clock     string
}

type ProcessWatch struct {
	PubChange chan WatchMessage
	Wg        *sync.WaitGroup
}
