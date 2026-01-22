package session

type Event int8
type Listener func(any)

var listeners = map[Event][]Listener{}

const (
	EventSessionNew     Event = iota //SESSION New,参数 *Data
	EventSessionCreated              //SESSION Create时,参数 *Data
	EventSessionRelease              //销毁SESSION时,参数 *Data
	EventHeartbeat                   //心跳，参数 心跳间隔 int32
)

func On(event Event, listener Listener) {
	listeners[event] = append(listeners[event], listener)
}

func Emit(event Event, value any) {
	ls, ok := listeners[event]
	if !ok {
		return
	}
	for _, l := range ls {
		l(value)
	}
}

func Listen(event Event, listener Listener) {
	On(event, listener)
}
