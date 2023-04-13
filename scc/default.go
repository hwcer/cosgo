package scc

import (
	"context"
	"time"
)

var scc = New(nil)

// GO 普通的GO
func GO(f func()) {
	scc.GO(f)
}

// CGO 带有取消通道的协程
func CGO(f handle) {
	scc.CGO(f)
}

// SGO 使用recover保护主进程,使用一个handle进行错误信息处理
func SGO(f handle) {
	scc.SGO(f)
}

func Try(f handle) {
	scc.Try(f)
}

func Add(delta int) {
	scc.WaitGroup.Add(delta)
}
func Done() {
	scc.WaitGroup.Done()
}

func Wait(timeout time.Duration) (err error) {
	return scc.Wait(timeout)
}

// Cancel ,callback:成功调用Close后 cancel之前调用
func Cancel() bool {
	return scc.Cancel()
}
func Context() context.Context {
	return scc.Context()
}

func Daemon(f handle) (cancel context.CancelFunc) {
	return scc.Daemon(f)
}

// Stopped 判断是否已经关闭
func Stopped() bool {
	return scc.Stopped()
}
