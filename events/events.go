package events

import "github.com/anthonyorona/logical_clock_sim/types"

type RankedEvent interface {
	GetTieBreaker() int
	GetLRank() int
	GetPID() types.ProcessID
	SetEventID(types.EventID)
	GetEventID() types.EventID
	GetEventType() string
}

type EventSequencer struct {
	sequence types.EventID
}

func (es *EventSequencer) SequenceEvent(re RankedEvent) {
	eventID := es.sequence
	re.SetEventID(eventID)
	es.sequence++
}
