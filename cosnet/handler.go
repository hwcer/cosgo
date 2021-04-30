package cosnet

import (
	"sync"
)

type HandlerEventType int32
type HandlerEventFunc func(Socket) bool
type HandlerMessageFunc func(Socket, *Message) bool

const (
	HandlerEventTypeConnect HandlerEventType = iota
	HandlerEventTypeDisconnect
)

type Handler interface {
	On(HandlerEventType, HandlerEventFunc)
	Emit(HandlerEventType, Socket) bool
	Message(Socket, *Message) bool //消息处理函数
}

type HandlerEvents struct {
	mutex    sync.Mutex
	event    []HandlerEventType
	callback []HandlerEventFunc
}

func (this *HandlerEvents) On(e HandlerEventType, f HandlerEventFunc) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.event = append(this.event, e)
	this.callback = append(this.callback, f)
}
func (this *HandlerEvents) Emit(e HandlerEventType, s Socket) bool {
	for i, k := range this.event {
		if k == e && !this.callback[i](s) {
			return false
		}
	}
	return true
}

type HandlerDefault struct {
	HandlerEvents
	Handle map[int]HandlerMessageFunc
}

func (r *HandlerDefault) Message(sock Socket, msg *Message) bool {
	if f, ok := r.Handle[int(msg.Head.Proto)]; ok {
		return f(sock, msg)
	}
	return true
}

func (r *HandlerDefault) Register(act int, fun HandlerMessageFunc) {
	if r.Handle == nil {
		r.Handle = map[int]HandlerMessageFunc{}
	}
	r.Handle[act] = fun
}
