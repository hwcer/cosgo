package await

import (
	"context"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/cosmo/cache"
	"sync/atomic"
	"time"
)

var (
	ErrTimeout    = values.Errorf(0, "timeout")
	ErrServerBusy = values.Errorf(1, "server busy,try again later")
)

func New() *Await {
	return &Await{}
}

type Await struct {
	c       chan *Message
	Timeout time.Duration
}

// Try 如果通道已满，立即放弃执行
func (this *Await) Try(handle cache.Handle, args any) (any, error) {
	msg := &Message{args: args, handle: handle, done: make(chan struct{})}
	select {
	case this.c <- msg:
	default:
		msg.err = ErrServerBusy
		close(msg.done)
	}
	return msg.Wait(this.Timeout)
}

// Call 同步调用handle并返回结果
func (this *Await) Call(handle cache.Handle, args any) (any, error) {
	msg := this.Sync(handle, args)
	return msg.Wait(this.Timeout)
}

// Sync 异步执行，不关心执行结果
// 也可以使用 Message.Done等待返回结果
func (this *Await) Sync(handle cache.Handle, args any) *Message {
	msg := &Message{args: args, handle: handle, done: make(chan struct{})}
	this.c <- msg
	return msg
}

func (this *Await) Start(cap int, timeout time.Duration) {
	this.c = make(chan *Message, cap)
	this.Timeout = timeout
	scc.CGO(this.process)
}

func (this *Await) process(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case m := <-this.c:
			this.handle(m)
		}
	}
}

func (this *Await) handle(msg *Message) {
	if !atomic.CompareAndSwapInt32(&msg.state, 0, 1) {
		return //对方等待超时已经放弃执行
	}
	defer func() {
		if e := recover(); e != nil {
			msg.err = values.Errorf(0, e)
		}
	}()
	msg.reply, msg.err = msg.handle(msg.args)
	close(msg.done)
}
