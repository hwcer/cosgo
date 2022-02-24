package utils

type EventsFunc func() error

func NewEmitter(breakOnError bool) *Emitter {
	return &Emitter{
		listener:     make(map[string][]EventsFunc),
		breakOnError: breakOnError,
	}
}

type Emitter struct {
	listener     map[string][]EventsFunc
	breakOnError bool //出现错误时是否终止执行后续监听器
}

func (this *Emitter) On(ename string, callback EventsFunc) {
	this.listener[ename] = append(this.listener[ename], callback)
}

func (this *Emitter) Emit(ename string) (errs []error) {
	funcs := this.listener[ename]
	if len(funcs) == 0 {
		return
	}
	for _, f := range funcs {
		if err := f(); err != nil {
			errs = append(errs, err)
			if this.breakOnError {
				break
			}
		}
	}
	return
}
