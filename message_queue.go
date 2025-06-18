package main

import "container/heap"

type Message struct {
	EventID     EventID
	LTime       LTime
	ProcessID   ProcessID
	MessageType MessageType
	Payload     map[string]int
}

type PriorityItem interface {
	GetLTime() int
	GetEventID() int
}

func (m *Message) GetLTime() int {
	return int(m.LTime)
}

func (m *Message) GetEventID() int {
	return int(m.EventID)
}

type EventQueue[T PriorityItem] []T

func (m *EventQueue[T]) Len() int { return len(*m) }

func (m *EventQueue[T]) Less(i, j int) bool {
	l := (*m)[i]
	r := (*m)[j]
	return l.GetLTime() < r.GetLTime() || (l.GetLTime() == r.GetLTime() && l.GetEventID() < r.GetEventID())
}

func (m *EventQueue[T]) Swap(i, j int) {
	(*m)[i], (*m)[j] = (*m)[j], (*m)[i]
}

func (m *EventQueue[T]) Push(v any) {
	tV, _ := v.(T)
	*m = append(*m, tV)
}

func (m *EventQueue[T]) Pop() any {
	old := *m
	n := len(old)
	item := old[n-1]
	*m = old[0 : n-1]
	return item
}

func (m *EventQueue[T]) RemoveAtIndices(indices map[int]struct{}) {
	var newQueue EventQueue[T]
	for i, e := range *m {
		if _, found := indices[i]; !found {
			newQueue = append(newQueue, e)
		}
	}
	*m = newQueue
	heap.Init(m)
}

func (m *EventQueue[T]) RemoveWhere(where func(T) bool) {
	var newQueue EventQueue[T]
	for _, element := range *m {
		if !where(element) {
			newQueue = append(newQueue, element)
		}
	}
	*m = newQueue
	heap.Init(m)
}
