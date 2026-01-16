package await

import (
	"sync/atomic"
	"time"
)

// 注意：Handle 类型在 await.go 中已经定义，这里不再重复定义

// Message 消息结构，用于在异步调用中传递参数和接收结果
type Message struct {
	err    error         // 错误信息
	args   any           // 传递给处理函数的参数
	done   chan struct{} // 完成信号通道
	state  int32         // 状态标记，用于原子操作
	reply  any           // 处理函数的返回值
	handle Handle        // 处理函数
}

// Wait 等待处理函数执行完成并返回结果
// @param t 超时时间
// @return any 处理函数的返回值
// @return error 错误信息，可能是ErrTimeout或处理函数返回的错误
func (this *Message) Wait(t time.Duration) (any, error) {
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

// Done 等待完成请求（已注释）
// timeout 等待超时时间,实际超时时间最大为 timeout * 2
//func (this *Message) Done(timeout time.Duration) (any, error) {
//	return this.Wait(timeout)
//}
