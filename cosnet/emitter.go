package cosnet

//事件必须在初始化时添加好，不能动态添加
type EventsFunc func(*Socket)

func NewEmitter(cap int) *Emitter {
	return &Emitter{
		listener: make([][]EventsFunc, cap),
	}
}

type Emitter struct {
	listener [][]EventsFunc
}

func (this *Emitter) On(e EventType, f EventsFunc) {
	this.listener[e] = append(this.listener[e], f)
}

func (this *Emitter) Emit(e EventType, s *Socket) {
	for _, f := range this.listener[e] {
		f(s)
	}
}
