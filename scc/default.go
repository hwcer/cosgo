package scc

import (
	"context"
	"time"
)

// Default 是SCC的默认实例，全局使用
var Default = New(nil)

// GO 启动一个普通的协程（使用默认实例）
// @param f 要执行的函数
func GO(f func()) {
	Default.GO(f)
}

// CGO 启动一个带有取消通道的协程（使用默认实例）
// @param f 要执行的处理函数，接收一个上下文参数
func CGO(f handle) {
	Default.CGO(f)
}

// SGO 启动一个使用recover保护的协程，防止主进程崩溃（使用默认实例）
// @param f 要执行的处理函数，接收一个上下文参数
func SGO(f handle) {
	Default.SGO(f)
}

// Try 尝试执行一个处理函数，使用recover捕获异常（使用默认实例）
// @param f 要执行的处理函数，接收一个上下文参数
func Try(f handle) {
	Default.Try(f)
}

// Add 向默认实例的WaitGroup添加计数
// @param delta 要添加的计数
func Add(delta int) {
	Default.WaitGroup.Add(delta)
}

// Done 减少默认实例的WaitGroup计数
func Done() {
	Default.WaitGroup.Done()
}

// Wait 阻塞模式等待所有协程结束（使用默认实例）
// @param timeout 超时时间，如果为0则无限等待
// @return error 错误信息，可能是超时错误
func Wait(timeout time.Duration) (err error) {
	return Default.Wait(timeout)
}

// Cancel 关闭所有协程（使用默认实例）
// @return bool 是否成功关闭，如果已经关闭过则返回false
func Cancel() bool {
	return Default.Cancel()
}

// Trigger 注册一个在服务器关闭时执行的函数（使用默认实例）
// @param handle 要执行的函数
func Trigger(handle func()) {
	Default.Trigger(handle)
}

// Stopped 判断默认实例是否已经关闭
// @return bool 是否已关闭
func Stopped() bool {
	return Default.Stopped()
}

// WithCancel 创建一个带有取消功能的子上下文（使用默认实例）
// @return context.Context 创建的子上下文
// @return context.CancelFunc 取消函数
func WithCancel() (context.Context, context.CancelFunc) {
	return Default.WithCancel()
}

// WithTimeout 创建一个带有超时功能的子上下文（使用默认实例）
// @param t 超时时间
// @return context.Context 创建的子上下文
// @return context.CancelFunc 取消函数
func WithTimeout(t time.Duration) (context.Context, context.CancelFunc) {
	return Default.WithTimeout(t)
}

// WithValue 创建一个带有键值对的子上下文（使用默认实例）
// @param parent 父上下文，如果为nil则使用默认实例的根上下文
// @param key 键
// @param val 值
// @return context.Context 创建的子上下文
func WithValue(parent context.Context, key, val any) context.Context {
	return Default.WithValue(parent, key, val)
}
