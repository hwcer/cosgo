package utils

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
	"time"
)

func NewSCC(ctx context.Context) *SCC {
	if ctx == nil {
		ctx = context.Background()
	}
	s := &SCC{WaitGroup: sync.WaitGroup{}}
	s.Context, s.cancel = context.WithCancel(ctx)
	s.WaitGroup.Add(1)
	return s
}

// 协程控制器
type SCC struct {
	stop      int32
	cancel    context.CancelFunc
	Context   context.Context
	WaitGroup sync.WaitGroup
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
func (s *SCC) CGO(f func(ctx context.Context)) {
	go func() {
		s.WaitGroup.Add(1)
		defer s.WaitGroup.Done()
		f(s.Context)
	}()
}

func (s *SCC) SGO(f func(ctx context.Context), handles ...TryHandle) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				if len(handles) == 0 {
					fmt.Printf("%v\n%v", err, string(debug.Stack()))
				} else {
					handles[0](err)
				}
			}
		}()
		s.WaitGroup.Add(1)
		defer s.WaitGroup.Done()
		f(s.Context)
	}()
}

func (s *SCC) Add(delta int) {
	s.WaitGroup.Add(delta)
}

func (s *SCC) Done() {
	s.WaitGroup.Done()
}

func (s *SCC) Wait(timeout time.Duration) (err error) {
	if timeout == 0 {
		s.WaitGroup.Wait()
		return
	}
	return Timeout(timeout, func() error {
		s.WaitGroup.Wait()
		return nil
	})
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

// Stopped 判断是否已经关闭
func (s *SCC) Stopped() bool {
	return s.stop != 0
}
