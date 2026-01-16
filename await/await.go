package await

import (
	"context"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/cosgo/values"
	"sync/atomic"
	"time"
)

// 错误定义
var (
	ErrTimeout    = values.Errorf(0, "timeout")    // 超时错误
	ErrServerBusy = values.Errorf(1, "server busy,try again later") // 服务器繁忙错误
)

// Handle 处理函数类型，用于处理异步任务
type Handle func(args any) (any, error)

// New 创建一个新的Await实例
// @param cap 通道容量，用于控制并发数
// @param timeout 默认超时时间
// @return *Await 返回创建的Await实例
func New(cap int, timeout time.Duration) *Await {
	r := &Await{}
	r.init(cap, timeout)
	return r
}

// Await 异步调用和等待机制的核心结构
type Await struct {
	c       chan *Message    // 消息通道，用于传递任务
	Timeout time.Duration    // 默认超时时间
}

// Try 尝试执行任务，如果通道已满，立即放弃执行
// @param handle 处理函数
// @param args 传递给处理函数的参数
// @return any 处理函数的返回值
// @return error 错误信息，可能是ErrServerBusy或处理函数返回的错误
func (this *Await) Try(handle Handle, args any) (any, error) {
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
// @param handle 处理函数
// @param args 传递给处理函数的参数
// @return any 处理函数的返回值
// @return error 错误信息，可能是ErrTimeout或处理函数返回的错误
func (this *Await) Call(handle Handle, args any) (any, error) {
	msg := this.Sync(handle, args)
	return msg.Wait(this.Timeout)
}

// Sync 异步执行，返回Message对象，可通过Message.Wait等待结果
// @param handle 处理函数
// @param args 传递给处理函数的参数
// @return *Message 返回Message对象，可用于等待执行结果
func (this *Await) Sync(handle Handle, args any) *Message {
	msg := &Message{args: args, handle: handle, done: make(chan struct{})}
	this.c <- msg
	return msg
}

// init 初始化Await实例
// @param cap 通道容量
// @param timeout 默认超时时间
func (this *Await) init(cap int, timeout time.Duration) {
	this.c = make(chan *Message, cap)
	this.Timeout = timeout
	scc.CGO(this.process)
}

// process 处理消息通道中的任务
// @param ctx 上下文，用于控制协程退出
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

// handle 处理单个消息
// @param msg 消息对象
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
