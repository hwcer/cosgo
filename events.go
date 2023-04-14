package cosgo

type EventType int32
type EventFunc func() error

const (
	EventTypStarting EventType = iota
	EventTypStarted
	EventTypStopping
	EventTypStopped
)

var events map[EventType][]EventFunc

func init() {
	events = make(map[EventType][]EventFunc)
}

func emit(e EventType) (err error) {
	if funcs, ok := events[e]; ok {
		for _, f := range funcs {
			if err = f(); err != nil {
				return
			}
		}
	}
	return
}

func On(e EventType, f EventFunc) {
	events[e] = append(events[e], f)
}
