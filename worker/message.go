package worker

import (
	"sync/atomic"
	"time"
)

type Handle func(args any) (reply any, err error)

type Message struct {
	err    error
	args   any
	done   chan struct{}
	state  int32
	reply  any
	handle Handle
}

func (this *Message) wait(t time.Duration) (any, error) {
	timer := time.NewTimer(t)
	defer timer.Stop()
	for {
		select {
		case <-this.done:
			return this.reply, this.err
		case <-timer.C:
			if !atomic.CompareAndSwapInt32(&this.state, 0, 1) {
				timer.Reset(t) //正在处理中
			} else {
				return nil, ErrTimeout
			}
		}
	}
}

// Done 等待完成请求
// timeout 等待超时时间,实际超时时间最大为 timeout * 2
func (this *Message) Done(timeout time.Duration) (any, error) {
	return this.wait(timeout)
}
