package process

import (
	"sync"

	"github.com/anthonyorona/distributed_sims/types"
)

type WatchMessage struct {
	ProcessID types.ProcessID
	State     string
	Clock     string
	QL        int
}

type ProcessWatch struct {
	PubChange chan WatchMessage
	Wg        *sync.WaitGroup
}
