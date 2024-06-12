package worker

import (
	"context"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/scc"
	"sync/atomic"
	"time"
)

var ErrTimeout = values.Errorf(0, "timeout")

func New(cap int) *Worker {
	i := &Worker{}
	i.c = make(chan *message, cap)
	return i
}

type Handle func(args any) error

type message struct {
	re     chan error
	args   any
	state  int32
	handle Handle
}

func (this *message) write(r error) {
	select {
	case this.re <- r:
	default:
	}
}

type Worker struct {
	c chan *message
}

func (this *Worker) Call(handle Handle, args any) error {
	msg := &message{args: args, handle: handle, re: make(chan error)}
	select {
	case this.c <- msg:
	default:
		return values.Errorf(0, "worker chan is full")
	}
	return this.wait(msg)
}

func (this *Worker) Start() {
	scc.CGO(this.process)
}

func (this *Worker) process(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-this.c:
			this.handle(m)
		}
	}
}

func (this *Worker) wait(msg *message) (re error) {
	var wait int32
	var doing bool
	timer := time.NewTimer(time.Second)
	defer timer.Stop()
	for {
		select {
		case re = <-msg.re:
			return
		case <-timer.C:
			if doing {
				if wait < 100 {
					wait += 1
					timer.Reset(time.Millisecond * 10) //正在处理中
				} else {
					return ErrTimeout
				}
			} else if !atomic.CompareAndSwapInt32(&msg.state, 0, 1) {
				wait += 1
				doing = true
				timer.Reset(time.Millisecond * 10) //正在处理中
			} else {
				return ErrTimeout
			}
		}
	}
}

func (this *Worker) handle(msg *message) {
	if !atomic.CompareAndSwapInt32(&msg.state, 0, 1) {
		return //对方等待超时已经放弃执行
	}
	var err error
	defer func() {
		msg.write(err)
	}()
	defer func() {
		if e := recover(); e != nil {
			err = values.Errorf(0, e)
		}
	}()
	err = msg.handle(msg.args)
}
