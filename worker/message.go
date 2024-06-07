package worker

type message struct {
	re     chan error
	args   any
	state  int32
	handle func(any) error
}

func (this *message) write(err error) {
	select {
	case this.re <- err:
	default:
	}
}
