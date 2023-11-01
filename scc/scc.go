package scc

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

type handle func(context.Context)
type daemon struct {
	handle  handle
	context context.Context
}

func catch(err error) {
	fmt.Printf("%v", err)
}

func New(ctx context.Context) *SCC {
	if ctx == nil {
		ctx = context.Background()
	}
	s := &SCC{Catch: catch, WaitGroup: sync.WaitGroup{}}
	s.context, s.cancel = context.WithCancel(ctx)
	s.WaitGroup.Add(1)
	return s
}

// SCC 协程控制器
type SCC struct {
	sync.WaitGroup
	stop    int32
	cancel  context.CancelFunc
	context context.Context
	Catch   func(error) //异常捕获,默认控制台打印
}

// GO 普通的GO
func (s *SCC) GO(f func()) {
	go func() {
		s.WaitGroup.Add(1)
		defer s.WaitGroup.Done()
		f()
	}()
}

// CGO 带有取消通道的协程
func (s *SCC) CGO(f handle) {
	go func() {
		s.WaitGroup.Add(1)
		defer s.WaitGroup.Done()
		f(s.context)
	}()
}

// SGO 使用recover保护主进程,使用一个handle进行错误信息处理
func (s *SCC) SGO(f handle) {
	go func() {
		s.Try(f)
	}()
}

func (s *SCC) Try(f handle) {
	defer func() {
		if e := recover(); e != nil {
			s.Catch(fmt.Errorf("%v\n%v", e, string(debug.Stack())))
		}
	}()
	s.WaitGroup.Add(1)
	defer s.WaitGroup.Done()
	f(s.context)
}

func (s *SCC) Wait(timeout time.Duration) (err error) {
	if timeout == 0 {
		s.WaitGroup.Wait()
	} else {
		err = s.Timeout(timeout, func() error {
			s.WaitGroup.Wait()
			return nil
		})
	}
	return
}

// Cancel ,callback:成功调用Close后 cancel之前调用
func (s *SCC) Cancel() bool {
	if !atomic.CompareAndSwapInt32(&s.stop, 0, 1) {
		return false
	}
	s.WaitGroup.Done()
	s.cancel()
	return true
}

func (s *SCC) Context() context.Context {
	return s.context
}

// Stopped 判断是否已经关闭
func (s *SCC) Stopped() bool {
	return s.stop > 0
}

// Daemon 守护协程,协程异常退出时会自动重启协程,一般使用在随主进程启动的固定协程
func (s *SCC) Daemon(f handle) (cancel context.CancelFunc) {
	d := &daemon{handle: f}
	d.context, cancel = context.WithCancel(s.context)
	s.daemon(d)
	return
}

func (s *SCC) daemon(d *daemon) {
	if s.Stopped() {
		return
	}
	select {
	case <-s.context.Done():
	case <-d.context.Done():
	default:
		go func() {
			defer func() {
				if e := recover(); e != nil {
					s.Catch(fmt.Errorf("%v\n%v", e, string(debug.Stack())))
				}
				s.daemon(d)
			}()
			s.WaitGroup.Add(1)
			defer s.WaitGroup.Done()
			d.handle(d.context)
		}()
	}
}
