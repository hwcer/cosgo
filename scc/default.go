package scc

import (
	"context"
	"time"
)

// Default 是全局默认 SCC 实例。
var Default = New(context.Background())

// GO 使用默认实例启动一个普通协程。
func GO(f func()) {
	Default.GO(f)
}

// CGO 使用默认实例启动一个带取消 context 的协程。
func CGO(f handle) {
	Default.CGO(f)
}

// SGO 使用默认实例启动一个带 recover 保护的协程。
func SGO(f handle) {
	Default.SGO(f)
}

// Try 使用默认实例同步执行 f,带 recover 保护。
func Try(f handle) {
	Default.Try(f)
}

// Add 向默认实例的 WaitGroup 添加计数。
func Add(delta int) {
	Default.WaitGroup.Add(delta)
}

// Done 向默认实例的 WaitGroup 递减计数。
func Done() {
	Default.WaitGroup.Done()
}

// Wait 使用默认实例阻塞等待所有协程结束,timeout=0 无限等待。
func Wait(timeout time.Duration) (err error) {
	return Default.Wait(timeout)
}

// Cancel 关闭默认实例的所有协程。
func Cancel() bool {
	return Default.Cancel()
}

// Trigger 向默认实例注册一个在关闭时执行的函数。
func Trigger(handle func()) {
	Default.Trigger(handle)
}

// Stopped 返回默认实例是否已关闭。
func Stopped() bool {
	return Default.Stopped()
}

// WithCancel 基于默认实例创建带取消的子 context。
func WithCancel() (context.Context, context.CancelFunc) {
	return Default.WithCancel()
}

// WithTimeout 基于默认实例创建带超时的子 context。
func WithTimeout(t time.Duration) (context.Context, context.CancelFunc) {
	return Default.WithTimeout(t)
}

// WithValue 基于默认实例的根 context 创建带键值对的子 context。
func WithValue(key, val any) context.Context {
	return Default.WithValue(key, val)
}
