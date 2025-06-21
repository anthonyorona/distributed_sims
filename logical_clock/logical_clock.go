package logical_clock

type LogicalClock[C interface{}, T interface{}] interface {
	GetLTime() C
	C(v T)
}
