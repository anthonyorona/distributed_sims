package main

import (
	"fmt"
	"sync"
	"time"
)

type ProcessID int
type EventID int
type MessageType int
type LTime int

type Sim struct {
	Request       *time.Ticker
	Usage         *time.Timer
	InternalEvent *time.Timer
}

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

type MessageDispatch struct {
	sequence EventID
}

func (md *MessageDispatch) NewMessage(pID ProcessID, lTime LTime, messageType MessageType, payload ...map[string]int) Message {
	EventID := md.sequence
	md.sequence += 1
	var p map[string]int
	if len(payload) == 1 {
		p = payload[0]
	}
	return Message{
		EventID:     EventID,
		ProcessID:   pID,
		LTime:       lTime,
		MessageType: messageType,
		Payload:     p,
	}
}

type WatchMessage struct {
	ProcessID
	Clock LTime
	State string
}

type ProcessWatch struct {
	PubChange chan WatchMessage
	Wg        *sync.WaitGroup
}

func (mt MessageType) String() string {
	switch mt {
	case Request:
		return "Request"
	case Ack:
		return "Ack"
	case Release:
		return "Release"
	case Internal:
		return "Internal"
	default:
		return fmt.Sprintf("UNKNOWN_MESSAGE_TYPE(%d)", mt)
	}
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
