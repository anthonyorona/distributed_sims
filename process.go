package main

import (
	"fmt"
	"math/rand"
	"time"
)

func (p *Process) C(events ...Message) {
	currentTime := p.Clock + 1
	if len(events) == 1 {
		lTime := events[0].LTime
		if lTime >= currentTime {
			currentTime = lTime + 1
		}
	}
	p.Clock = currentTime
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
		if p.ID != message.ProcessID {
			d.RecvChan <- message
		}
	}
}

func (p *Process) Request() {
	p.C()
	request := Message{
		ProcessID:   p.ID,
		LTime:       p.Clock,
		MessageType: Request,
	}
	p.MQueue = append(p.MQueue, request)
	p.SetSRState(Requested)
	p.Broadcast(request)
}

func (p *Process) Release() {
	p.C()
	if p.ResourceHold == nil {
		panic("This should never happen")
	}
	p.MQueue = RemoveWhere(p.MQueue, func(m Message) bool {
		return p.ResourceHold.ProcessID == m.ProcessID && p.ResourceHold.LTime == m.LTime
	})
	p.Broadcast(Message{
		ProcessID:   p.ID,
		LTime:       p.Clock,
		MessageType: Release,
		Payload: map[string]int{
			"ProcessID": int(p.ID),
			"LTime":     int(p.ResourceHold.LTime),
		},
	})
	p.ResourceHold = nil
	p.SetSRState(Free)
}

// When process p receives message Tm:pi requests it places it on its request queue and sends timestamped ack message to Pi
func (p *Process) Receive(message Message) {
	p.C()
	p.MQueue = append(p.MQueue, message)
	if message.MessageType == Request {
		// response
		p.Directory[message.ProcessID].RecvChan <- Message{
			ProcessID:   p.ID,
			LTime:       p.Clock,
			MessageType: Ack,
		}
	} else if message.MessageType == Ack {
		p.AckSet[message.ProcessID] = struct{}{}
		if len(p.AckSet) == len(p.Directory) {
			p.SetSRState(Acknowledged)
			p.AckSet = make(map[ProcessID]struct{})
		}
	} else if message.MessageType == Release {
		p.MQueue = RemoveWhere(p.MQueue, func(m Message) bool {
			return m.Payload["ProcessID"] == int(m.ProcessID) && m.Payload["LTime"] == int(m.LTime)
		})
	} else {
		panic(fmt.Sprintf("Message type %d not recognized", message.MessageType))
	}

	if p.MQueue[0].ProcessID == p.ID {
		matches := make(map[ProcessID]int)
		for _, peer := range p.Directory {
			matches[peer.ProcessID] = -1
			for i, queuedMessage := range p.MQueue {
				if peer.ProcessID == queuedMessage.ProcessID && p.MQueue[0].LTime < queuedMessage.LTime {
					matches[peer.ProcessID] = i
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
		p.ResourceHold = &p.MQueue[0]
		p.Sim.Usage = time.NewTimer(time.Duration(rand.Intn(1000)+500) * time.Millisecond)
		p.MQueue = RemoveAtIndices(p.MQueue, indicesToRemoveSet)
		p.SetSRState(Holding)
	}
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
			if rand.New(rand.NewSource(time.Now().UnixNano())).Intn(2) == 0 {
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
