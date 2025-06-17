package main

import (
	"context"
	"sync"
	"time"
)

type ProcessID int
type EventID int
type MessageType int
type LTime int
type MessageQueue []*Message

type DirectoryEntry struct {
	ProcessID ProcessID
	RecvChan  chan Message
}

const (
	Request MessageType = iota
	Ack
	Release
	Internal
)

const (
	Free int = iota
	Requested
	Acknowledged
	Holding
)

type WatchMessage struct {
	ProcessID
	Clock LTime
	State string
}

type ProcessWatch struct {
	PubChange chan WatchMessage
	Wg        *sync.WaitGroup
}

type Process struct {
	ID        ProcessID
	Ctx       context.Context
	Clock     LTime
	Directory []DirectoryEntry
	AckSet    map[ProcessID]struct{}
	Sim
	RecvChan     chan Message
	MQueue       []Message
	ResourceHold *Message
	ProcessWatch
	SRState int
}

type Sim struct {
	Request       *time.Ticker
	Usage         *time.Timer
	InternalEvent *time.Timer
}

func (p *Process) GetSRState() string {
	switch p.SRState {
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

type Message struct {
	EventID     EventID
	LTime       LTime
	ProcessID   ProcessID
	MessageType MessageType
	Payload     map[string]int
}

func (m MessageQueue) Len() int { return len(m) }

func (m MessageQueue) Less(i, j int) bool {
	l := m[i]
	r := m[j]
	return l.LTime < r.LTime || (l.LTime == r.LTime && l.EventID < r.EventID)
}

func (m MessageQueue) Swap(i, j int) {
	m[i], m[j] = m[j], m[i]
}

func (m *MessageQueue) Push(message *Message) {
	*m = append(*m, message)
}

func (m *MessageQueue) Pop() *Message {
	old := *m
	n := len(old)
	message := old[n-1]
	old[n-1] = nil
	*m = old[0 : n-1]
	return message
}
