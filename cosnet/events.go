package cosnet

import (
	"sync"
)

type EventsType int32
type EventsFunc func(Socket)

const (
	EventsTypeConnect EventsType = iota
	EventsTypeDisconnect
)

func NewEvents() *Events {
	return &Events{}
}

type Events struct {
	mutex    sync.Mutex
	event    []EventsType
	callback []EventsFunc
}

func (this *Events) On(e EventsType, f EventsFunc) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.event = append(this.event, e)
	this.callback = append(this.callback, f)
}

func (this *Events) Emit(e EventsType, s Socket) {
	for i, k := range this.event {
		if k == e {
			this.callback[i](s)
		}
	}
}
