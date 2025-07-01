package events

import (
	"container/heap"
	"fmt"
)

type EventQueue[T RankedEvent] []T

func (m *EventQueue[T]) Len() int { return len(*m) }

func (m *EventQueue[T]) Less(i, j int) bool {
	l := (*m)[i]
	r := (*m)[j]
	return l.GetLRank() < r.GetLRank() || (l.GetTieBreaker() == r.GetTieBreaker() && l.GetEventID() < r.GetEventID())
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

func (m *EventQueue[T]) PrettyPrint(label string) {
	fmt.Printf("--- EventQueue Contents (%s) ---\n", label)
	if m.Len() == 0 {
		fmt.Println("  (Queue is empty)")
		fmt.Println("---------------------------------")
		return
	}

	for i, item := range *m { // Keep your original iterator
		// THIS IS THE ONLY SIGNIFICANT ADDITION: Check for nil before calling methods
		if fmt.Sprintf("%v", item) == "<nil>" { // General way to check if 'item' is nil or an interface holding nil
			fmt.Printf("  [%d] <nil item>\n", i)
			continue // Skip this item
		}

		// If we get here, 'item' is not nil, so it's safe to call its methods
		fmt.Printf("  [%d] LRank: %d, EventID: %d, EventType: %s, Tie: %d\n",
			i, item.GetLRank(), item.GetEventID(), item.GetEventType(), item.GetTieBreaker())
	}
	fmt.Println("---------------------------------")
}
