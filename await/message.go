package await

import (
	"sync/atomic"
	"time"
)

// Message 异步任务的执行上下文和结果持有者
type Message struct {
	err    error
	args   any
	done   chan struct{}
	state  int32
	reply  any
	handle Handle
}

// Wait 等待任务完成并返回结果
// 快路径：任务已完成时（如 Try 被拒绝）直接返回，不分配 timer
// 超时说明：若 handler 正在执行中（state=1），timer 会重置再等一轮，
// 实际最大等待时间为 timeout × 2
func (this *Message) Wait(t time.Duration) (any, error) {
	// 快路径：任务已完成（Try 被拒或任务极速返回），跳过 timer 分配
	select {
	case <-this.done:
		return this.reply, this.err
	default:
	}
	// 慢路径：分配 timer 等待
	timer := time.NewTimer(t)
	defer timer.Stop()
	for {
		select {
		case <-this.done:
			return this.reply, this.err
		case <-timer.C:
			// CAS(0→1)：尝试标记为"已放弃"
			// 成功 → handler 尚未开始，直接超时返回
			// 失败 → handler 正在执行（state 已被 handler CAS 为 1），再等一轮
			if !atomic.CompareAndSwapInt32(&this.state, 0, 1) {
				timer.Reset(t)
			} else {
				return nil, ErrTimeout
			}
		}
	}
}
