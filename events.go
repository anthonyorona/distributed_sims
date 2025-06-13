package main

type EventID int
type MessageType int

const (
	RequestEvent MessageType = iota
	AckEvent
	ReleaseEvent
	InternalEvent
)

const (
	PrivateResource ResourceID = iota
	SharedResource
)

type Event interface {
	ID() EventID
	PID() ProcessID
	Time() Timestamp
	MType() MessageType
}

type Message struct {
	EventID     EventID
	Timestamp   Timestamp
	ProcessID   ProcessID
	MessageType MessageType
}

func (m *Message) ID() EventID {
	return m.EventID
}

func (m *Message) PID() ProcessID {
	return m.ProcessID
}

func (m *Message) Time() Timestamp {
	return m.Timestamp
}

func (m *Message) MType() MessageType {
	return m.MessageType
}
