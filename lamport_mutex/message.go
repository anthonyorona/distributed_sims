package lamport_mutex

import (
	"fmt"

	"github.com/anthonyorona/distributed_sims/types"
)

type MessageType int

const (
	Request MessageType = iota
	Ack
	Release
	Internal
)

type Message struct {
	ProcessID   types.ProcessID
	EventID     types.EventID
	LTime       LamportTimeStamp
	MessageType MessageType
	Payload     map[string]int
}

func (m *Message) GetLTime() LamportTimeStamp {
	return m.LTime
}

func (m *Message) GetEventID() types.EventID {
	return m.EventID
}

func (m *Message) SetEventID(eventID types.EventID) {
	m.EventID = eventID
}

func (m *Message) GetPID() types.ProcessID {
	return m.ProcessID
}

func (m *Message) GetEventType() string {
	return m.MessageType.String()
}

func (m *Message) GetLRank() int {
	return int(m.LTime)
}

func (m *Message) GetTieBreaker() int {
	return int(m.GetPID())
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
