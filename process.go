package main

import (
	"context"
	"math/rand"
	"sort"
	"sync"
	"time"
)

type Timestamp int
type ProcessID int
type ResourceID int

type Process struct {
	ID                 ProcessID
	Broadcaster        *Broadcaster
	Directory          []*Process
	ResourceHold       Event
	RecvChan           chan Event
	SendChan           chan Event
	EventQueue         []Event
	Ctx                context.Context
	Clock              Timestamp
	DoneWg             *sync.WaitGroup
	acksNeeded         int
	RequestTicker      *time.Ticker
	SimulateUsage      *time.Timer
	InternalEventTimer *time.Timer
}

// Clock Condition: For any events a, b if a -> b then C(a) < C(b)
// Each Process has a clock that assigns a number to a event
func (p *Process) C(events ...Event) {
	// If a and b are events in Process Pi and a comes before b
	// OR
	// If a is the sending of a message by process Pi and b is
	// is the receipt of that message by process Pj then
	// then Ci<a> < Ci<b>
	currentTime := p.Clock + 1
	if len(events) == 1 {
		eventTime := events[0].Time()
		if eventTime >= currentTime {
			currentTime = eventTime + 1
		}
	}
	p.Clock = currentTime
}

// To request a resource process p broadcasts the message Tm:Pi requests resource and, puts that message on its Queue
// Duplicate requests are not allowed
func (p *Process) Request() {
	if p.acksNeeded < 1 {
		p.C()
		request := Message{
			ProcessID:   p.ID,
			Timestamp:   p.Clock,
			MessageType: RequestEvent,
		}
		p.EventQueue = append(p.EventQueue, &request)
		p.acksNeeded = len(p.Directory)
		p.Broadcaster.Send(&request)
	}
}

// When process p receives message Tm:pi requests it places it on its request queue and sends timestamped ack message to Pi
func (p *Process) Receive(event Event) {
	p.C()
	p.EventQueue = append(p.EventQueue, event)
	if event.MType() == RequestEvent {
		p.Directory[event.PID()].RecvChan <- &Message{
			ProcessID:   p.ID,
			Timestamp:   p.Clock,
			MessageType: AckEvent,
		}
	}
	sort.Slice(p.EventQueue, func(i, j int) bool {
		return p.EventQueue[i].Time() < p.EventQueue[j].Time() ||
			p.EventQueue[i].Time() == p.EventQueue[j].Time() && p.EventQueue[i].ID() < p.EventQueue[j].ID()
	})
	/*
		Process P/is granted the resource when the follow-
		ing two conditions are satisfied: (i) There is a Tm:Pi
		requests resource message in its request queue which is
		ordered before any other request in its queue by the
		relation ~. (To define the relation "~" for messages,
		we identify a message with the event of sending it.) (ii)
		P~ has received a message from every other process time-
		stamped later than Tin.
	*/
	firstEvent := p.EventQueue[0].PID()
	firstEventTime := p.EventQueue[0].Time()
	if firstEvent == p.ID {
		matches := make(map[ProcessID]int)
		for _, peer := range p.Directory {
			matches[peer.ID] = -1
			for i, queuedEvent := range p.EventQueue {
				if peer.ID == queuedEvent.PID() && firstEventTime < queuedEvent.Time() {
					matches[peer.ID] = i
					break
				}
			}
		}
		acquired := true
		indicesToRemoveSet := map[int]struct{}{0: {}}
		for _, v := range matches {
			if v == -1 {
				acquired = false
				break
			}
			indicesToRemoveSet[v] = struct{}{}
		}
		if !acquired {
			return
		}
		p.ResourceHold = p.EventQueue[0]
		var updatedQueue []Event
		for i, _ := range indicesToRemoveSet {
			if _, found := indicesToRemoveSet[i]; !found {
				updatedQueue = append(updatedQueue, p.EventQueue[i])
			}
		}
		p.EventQueue = updatedQueue
	}
}

// To release resource, p removes any Tm:Pi requests resource message and sends Pi releases resource message Waitn round amojutn of time then release
func (p *Process) Release() {
	p.C()
	if p.ResourceHold == nil {
		panic("This should never happen")
		return
	}
	// TODO
	release := Message{
		ProcessID:   p.ID,
		Timestamp:   p.Clock,
		MessageType: ReleaseEvent,
	}
	p.Broadcaster.Send(&release)
}

func (p *Process) InternalEvent() {
	p.C() // just move the clock forward
}

func (p *Process) PRoutine() {
	p.RequestTicker = time.NewTicker(1 * time.Second)
	p.InternalEventTimer = time.NewTimer(time.Duration(rand.Intn(151)+100) * time.Millisecond)
	for {
		select {
		case event := <-p.RecvChan:
			p.Receive(event)
		// Every second, send a message to claim resource with probability 1/2 if no request pending
		case <-p.RequestTicker.C:
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			if r.Intn(2) == 0 {
				p.Request()
			}
		case <-p.InternalEventTimer.C:
			p.InternalEventTimer.Reset(time.Duration(rand.Intn(151)+100) * time.Millisecond)
			p.InternalEvent()
		case <-p.SimulateUsage.C:
			p.SimulateUsage = nil
			p.Release()
		case <-p.Ctx.Done():
			p.RequestTicker.Stop()
			p.InternalEventTimer.Stop()
			if p.SimulateUsage != nil {
				p.SimulateUsage.Stop()
			}
			p.DoneWg.Done()
			return
		}
	}
}
