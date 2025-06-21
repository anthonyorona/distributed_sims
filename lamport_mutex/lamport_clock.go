package lamport_mutex

import "strconv"

type LamportTimeStamp int

type LamportClock struct {
	time LamportTimeStamp
}

func NewLamportClock() LamportClock {
	return LamportClock{
		time: LamportTimeStamp(1),
	}
}

func (l *LamportClock) GetLTime() LamportTimeStamp {
	return l.time
}

func (l *LamportClock) C(events ...*Message) {
	lTime := l.GetLTime()
	if len(events) == 1 {
		recvLTime := events[0].GetLTime()
		if recvLTime > lTime {
			l.time = recvLTime + 1
			return
		}
	}
	l.time++
}

func (l *LamportClock) String() string {
	return strconv.Itoa(int(l.GetLTime()))
}
