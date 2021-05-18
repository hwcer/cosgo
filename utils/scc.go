package utils

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

func NewSCC(ctx context.Context) *SCC {
	if ctx == nil {
		ctx = context.Background()
	}
	s := &SCC{WaitGroup: sync.WaitGroup{}}
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.Add(1)
	return s
}

//协程控制器
type SCC struct {
	sync.WaitGroup
	ctx    context.Context
	stop   int32
	cancel context.CancelFunc
}

//GO 普通的GO
func (s *SCC) GO(f func()) {
	go func() {
		s.Add(1)
		defer s.Done()
		f()
	}()
}

//CGO 带有取消通道的协程
func (s *SCC) CGO(f func(ctx context.Context)) {
	go func() {
		s.Add(1)
		defer s.Done()
		f(s.ctx)
	}()
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

//Close ,callback:成功调用Close后 cancel之前调用
func (s *SCC) Close() bool {
	if !atomic.CompareAndSwapInt32(&s.stop, 0, 1) {
		return false
	}
	s.cancel()
	s.Done()
	return true
}

//判断是否已经关闭
func (s *SCC) Stopped() bool {
	select {
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}
func (s *SCC) Cancel() {
	s.cancel()
}

func (s *SCC) Context() context.Context {
	return s.ctx
}

func Timeout(d time.Duration, fn func() error) error {
	cher := make(chan error)
	go func() {
		cher <- fn()
	}()
	select {
	case err := <-cher:
		return err
	case <-time.After(d):
		return ErrorTimeout
	}
}
