package lamport_mutex

import (
	"container/heap"
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/anthonyorona/logical_clock_sim/common"
	"github.com/anthonyorona/logical_clock_sim/events"
	"github.com/anthonyorona/logical_clock_sim/process"
	"github.com/anthonyorona/logical_clock_sim/types"
)

// Process states for the Lamport distributed mutex simulation
const (
	Free types.PState = iota
	Requested
	Acknowledged
	Holding
)

type LamportProcess struct {
	process.Process[*Message]
	LamportClock
	LamportSim
	AckSet       map[types.ProcessID]struct{}
	ResourceHold *Message
	Outstanding  *Message
}

type LamportSim struct {
	Request       *time.Ticker
	Usage         *time.Timer
	InternalEvent *time.Timer
}

func NewProcess(ctx context.Context, pid types.ProcessID, initialMessage Message, recvChan chan *Message, pw process.ProcessWatch) LamportProcess {
	mQueue := make(events.EventQueue[*Message], 0)
	heap.Init(&mQueue)
	mCopy := initialMessage
	heap.Push(&mQueue, &mCopy)
	return LamportProcess{
		Process: process.Process[*Message]{
			Ctx:          ctx,
			ID:           pid,
			RecvChan:     recvChan,
			EventQueue:   mQueue,
			ProcessWatch: pw,
			PState:       Free,
		},
		LamportClock: NewLamportClock(),
		AckSet:       make(map[types.ProcessID]struct{}),
	}
}

func (p *LamportProcess) SetPState(v types.PState) {
	p.PState = v
	p.ProcessWatch.PubChange <- process.WatchMessage{
		ProcessID: p.ID,
		Clock:     p.LamportClock.String(),
		State:     p.GetPState(),
	}
}

func (p *LamportProcess) GetPState() string {
	switch p.PState {
	case Free:
		return "Free"
	case Requested:
		return "Requested"
	case Acknowledged:
		return "Acknowledged"
	case Holding:
		return "Holding"
	default:
		return "Undefined"
	}
}

func (p *LamportProcess) NewMessage(pID types.ProcessID, lTime LamportTimeStamp, messageType MessageType, opt ...map[string]int) Message {
	var payload map[string]int
	if len(opt) == 1 {
		payload = opt[0]
	}
	m := Message{
		ProcessID:   pID,
		LTime:       lTime,
		MessageType: messageType,
		Payload:     payload,
	}
	p.EventSequencer.SequenceEvent(&m)
	return m
}

func (p *LamportProcess) Done() {
	p.LamportSim.Request.Stop()
	p.LamportSim.InternalEvent.Stop()
	if p.LamportSim.Usage != nil {
		p.LamportSim.Usage.Stop()
	}
	close(p.RecvChan)
	p.ProcessWatch.Wg.Done()
}

func (p *LamportProcess) InternalEvent() {
	p.C() // just move the clock forward
}

func (p *LamportProcess) Request() {
	if p.PState == Free {
		p.C()
		request := p.NewMessage(p.ID, p.GetLTime(), Request)
		p.Outstanding = &request
		heap.Push(&p.EventQueue, &request)
		p.SetPState(Requested)
		p.AckSet = make(map[types.ProcessID]struct{})
		p.Broadcast(&request)
	}
}

func (p *LamportProcess) Release() {
	p.C()
	if p.ResourceHold == nil {
		panic("This should never happen")
	}
	p.EventQueue.RemoveWhere(func(m *Message) bool {
		return p.ResourceHold.GetPID() == m.GetPID() && p.ResourceHold.GetLRank() == m.GetLRank()
	})
	p.SetPState(Free)
	m := p.NewMessage(p.ID, p.GetLTime(), Release, map[string]int{
		"ProcessID": int(p.ID),
		"LRank":     int(p.ResourceHold.GetLRank()),
	})
	p.Broadcast(&m)
	p.ResourceHold = nil
}

// When process p receives message Tm:pi requests it places it on its request queue and sends timestamped ack message to Pi
func (p *LamportProcess) Receive(message *Message) {
	p.C(message)
	heap.Push(&p.EventQueue, &message)
	if message.MessageType == Request {
		m := p.NewMessage(p.ID, p.GetLTime(), Ack)
		p.Directory[message.ProcessID].RecvChan <- &m
	} else if message.MessageType == Ack {
		p.AckSet[message.ProcessID] = struct{}{}
	} else if message.MessageType == Release {
		p.EventQueue.RemoveWhere(func(m *Message) bool {
			return message.Payload["ProcessID"] == int(m.GetPID()) &&
				message.Payload["LRank"] == int(m.GetLRank())
		})
	} else {
		panic(fmt.Sprintf("Message type %d not recognized", message.MessageType))
	}

	if p.PState == Requested && p.Outstanding != nil {
		var smallest *Message
		for _, m := range p.EventQueue {
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
			smallest.GetPID() == p.Outstanding.GetPID() &&
			smallest.GetLRank() == p.Outstanding.GetLRank() {
			match = true
		}

		acked := len(p.AckSet) == (len(p.Directory) - 1)
		if match && acked {
			p.ResourceHold = p.Outstanding
			p.SetPState(Holding)
			p.EventQueue.RemoveWhere(func(m *Message) bool {
				return m.GetPID() == p.ID && m.GetLRank() == p.ResourceHold.GetLRank() && m.MessageType == Request
			})
			p.LamportSim.Usage = time.NewTimer(common.GetRandomDuration(1000, 500))
			p.AckSet = make(map[types.ProcessID]struct{})
			p.Outstanding = nil
		}
	}
	// p.EventQueue.PrettyPrint(fmt.Sprintf("PROCESS ID: %d, SR State: %s", p.ID, p.GetSRState()))
}

func (p *LamportProcess) Start() {
	p.LamportSim.Request = time.NewTicker(1 * time.Second)
	p.LamportSim.InternalEvent = time.NewTimer(common.GetRandomDuration(1000, 500))
	var usageChan <-chan time.Time
	for {
		if p.LamportSim.Usage != nil {
			usageChan = p.LamportSim.Usage.C
		} else {
			usageChan = nil // Ensure it's nil if Usage was cleared
		}
		select {
		case m := <-p.RecvChan:
			p.Receive(m)
		case <-p.LamportSim.Request.C:
			if rand.IntN(2) == 0 {
				p.Request()
			}
		case <-p.LamportSim.InternalEvent.C:
			p.LamportSim.InternalEvent.Reset(common.GetRandomDuration(1000, 500))
			p.InternalEvent()
		case <-usageChan:
			p.LamportSim.Usage = nil
			p.Release()
		case <-p.Ctx.Done():
			p.Done()
			return
		}
	}
}
