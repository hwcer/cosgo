package scc

import (
	"context"
	"time"
)

var Default = New(nil)

// GO 普通的GO
func GO(f func()) {
	Default.GO(f)
}

// CGO 带有取消通道的协程
func CGO(f handle) {
	Default.CGO(f)
}

// SGO 使用recover保护主进程,使用一个handle进行错误信息处理
func SGO(f handle) {
	Default.SGO(f)
}

func Try(f handle) {
	Default.Try(f)
}

func Add(delta int) {
	Default.WaitGroup.Add(delta)
}
func Done() {
	Default.WaitGroup.Done()
}

func Wait(timeout time.Duration) (err error) {
	return Default.Wait(timeout)
}

// Cancel ,callback:成功调用Close后 cancel之前调用
func Cancel() bool {
	return Default.Cancel()
}

// Stopped 判断是否已经关闭
func Stopped() bool {
	return Default.Stopped()
}

func WithCancel() (context.Context, context.CancelFunc) {
	return Default.WithCancel()
}

func WithTimeout(t time.Duration) (context.Context, context.CancelFunc) {
	return Default.WithTimeout(t)
}
