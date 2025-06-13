package main

import (
	"context"
	"math/rand"
	"sync"
	"time"
)

type Broadcaster struct {
	Mutex *sync.Mutex
	Queue []Event
}

// Messages are received in same order as they are sent
func (b *Broadcaster) Send(event Event) {
	b.Mutex.Lock()
	defer b.Mutex.Unlock()
	b.Queue = append(b.Queue, event)
}

// All messages are eventually received (using delay and acknowledgement)
func (b *Broadcaster) Start(ctx context.Context, processes []*Process, wg *sync.WaitGroup) {
	wg.Add(1)
	for {
		// Pause between iterations
		jitterMilliseconds := ((rand.Float64() * 2.0) - 1.0) * 1000
		jitterDuration := time.Duration(jitterMilliseconds * float64(time.Millisecond))
		totalDelay := 1*time.Second + jitterDuration
		timer := time.NewTimer(totalDelay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			wg.Done()
			return
		}
		eventFound := false
		var event Event
		b.Mutex.Lock()
		if len(b.Queue) > 0 {
			event = b.Queue[0]
			b.Queue = b.Queue[1:]
			eventFound = true
		}
		b.Mutex.Unlock()
		if eventFound {
			sender := event.PID()
			for _, p := range processes {
				if p.ID != sender {
					go func(ch chan Event, event Event) {
						ch <- event
					}(p.RecvChan, event)
				}
			}
		}
	}
}
