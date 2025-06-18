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
	MQueue       EventQueue[*Message]
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
