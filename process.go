package main

import (
	"container/heap"
	"context"
	"fmt"
	"math/rand/v2"
	"time"
)

type Process struct {
	ID        ProcessID
	Ctx       context.Context
	Clock     LTime
	Directory []DirectoryEntry
	AckSet    map[ProcessID]struct{}
	Sim
	RecvChan     chan Message
	MQueue       EventQueue[*Message]
	ResourceHold *Message
	Outstanding  *Message
	ProcessWatch
	MessageDispatch
	SRState int
	rng     *rand.Rand
}

func (p *Process) C(events ...Message) {
	if len(events) == 1 {
		recvLTime := events[0].LTime
		if recvLTime > p.Clock {
			p.Clock = recvLTime + 1
			return
		}
	}
	p.Clock++
}

func (p *Process) SetSRState(v int) {
	p.SRState = v
	p.ProcessWatch.PubChange <- WatchMessage{
		ProcessID: p.ID,
		Clock:     p.Clock,
		State:     p.GetSRState(),
	}
}

func (p *Process) Broadcast(message Message) {
	time.Sleep(getRandomDuration(200, 100))
	for _, d := range p.Directory {
		if d.ProcessID != message.ProcessID {
			d.RecvChan <- message
		}
	}
}

func (p *Process) Request() {
	if p.SRState == Free {
		p.C()
		request := p.MessageDispatch.NewMessage(p.ID, p.Clock, Request)
		p.Outstanding = &request
		heap.Push(&p.MQueue, &request)
		p.SetSRState(Requested)
		p.AckSet = make(map[ProcessID]struct{})
		p.Broadcast(request)
	}
}

func (p *Process) Release() {
	p.C()
	if p.ResourceHold == nil {
		panic("This should never happen")
	}
	p.MQueue.RemoveWhere(func(m *Message) bool {
		return p.ResourceHold.ProcessID == m.ProcessID && p.ResourceHold.LTime == m.LTime
	})
	p.SetSRState(Free)
	p.Broadcast(p.MessageDispatch.NewMessage(p.ID, p.Clock, Release, map[string]int{
		"ProcessID": int(p.ID),
		"LTime":     int(p.ResourceHold.LTime),
	}))
	p.ResourceHold = nil
}

// When process p receives message Tm:pi requests it places it on its request queue and sends timestamped ack message to Pi
func (p *Process) Receive(message Message) {
	p.C(message)
	heap.Push(&p.MQueue, &message)
	if message.MessageType == Request {
		p.Directory[message.ProcessID].RecvChan <- p.MessageDispatch.NewMessage(p.ID, p.Clock, Ack)
	} else if message.MessageType == Ack {
		p.AckSet[message.ProcessID] = struct{}{}
	} else if message.MessageType == Release {
		p.MQueue.RemoveWhere(func(m *Message) bool {
			return message.Payload["ProcessID"] == int(m.ProcessID) &&
				message.Payload["LTime"] == int(m.LTime)
		})
	} else {
		panic(fmt.Sprintf("Message type %d not recognized", message.MessageType))
	}

	if p.SRState == Requested && p.Outstanding != nil {
		var smallest *Message
		for _, m := range p.MQueue {
			if m.MessageType == Request {
				if smallest == nil ||
					m.LTime < smallest.LTime ||
					(m.LTime == smallest.LTime && m.ProcessID < smallest.ProcessID) {
					smallest = m
				}
			}
		}

		match := false
		if smallest != nil &&
			smallest.ProcessID == p.Outstanding.ProcessID &&
			smallest.LTime == p.Outstanding.LTime {
			match = true
		}

		acked := len(p.AckSet) == (len(p.Directory) - 1)
		if match && acked {
			p.ResourceHold = p.Outstanding
			p.SetSRState(Holding)
			p.MQueue.RemoveWhere(func(m *Message) bool {
				return m.ProcessID == p.ID && m.LTime == p.ResourceHold.LTime && m.MessageType == Request
			})
			p.Sim.Usage = time.NewTimer(getRandomDuration(1000, 500))
			p.AckSet = make(map[ProcessID]struct{})
			p.Outstanding = nil
		}
	}
	// p.MQueue.PrettyPrint(fmt.Sprintf("PROCESS ID: %d, SR State: %s", p.ID, p.GetSRState()))
}

func (p *Process) InternalEvent() {
	p.C() // just move the clock forward
}

func (p *Process) Start() {
	p.Sim.Request = time.NewTicker(1 * time.Second)
	p.Sim.InternalEvent = time.NewTimer(getRandomDuration(1000, 500))
	var usageChan <-chan time.Time
	for {
		if p.Sim.Usage != nil {
			usageChan = p.Sim.Usage.C
		} else {
			usageChan = nil // Ensure it's nil if Usage was cleared
		}
		select {
		case message := <-p.RecvChan:
			p.Receive(message)
		case <-p.Sim.Request.C:
			if rand.IntN(2) == 0 {
				p.Request()
			}
		case <-p.Sim.InternalEvent.C:
			p.Sim.InternalEvent.Reset(getRandomDuration(1000, 500))
			p.InternalEvent()
		case <-usageChan:
			p.Sim.Usage = nil
			p.Release()
		case <-p.Ctx.Done():
			p.Done()
			return
		}
	}
}

func (p *Process) Done() {
	p.Sim.Request.Stop()
	p.Sim.InternalEvent.Stop()
	if p.Sim.Usage != nil {
		p.Sim.Usage.Stop()
	}
	close(p.RecvChan)
	p.ProcessWatch.Wg.Done()
}
