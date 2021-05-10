package utils

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

func NewSCC(ctx context.Context) *SCC {
	if ctx == nil {
		ctx = context.Background()
	}
	s := &SCC{}
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.wgp.Add(1)
	return s
}

//协程控制器
type SCC struct {
	wgp    sync.WaitGroup
	ctx    context.Context
	stop   int32
	cancel context.CancelFunc
}

//GO 普通的GO
func (s *SCC) GO(f func()) {
	go func() {
		s.wgp.Add(1)
		defer s.wgp.Done()
		f()
	}()
}

//CGO 带有取消通道的协程
func (s *SCC) CGO(f func(ctx context.Context)) {
	go func() {
		s.wgp.Add(1)
		defer s.wgp.Done()
		f(s.ctx)
	}()
}

func (s *SCC) Ctx() context.Context {
	return s.ctx
}

func (s *SCC) Wait() {
	s.wgp.Wait()
}

//Close
func (s *SCC) Close() error {
	if !atomic.CompareAndSwapInt32(&s.stop, 0, 1) {
		return errors.New("SCC Close Exist")
	}
	s.wgp.Done()
	s.cancel()
	stopChan := make(chan bool)
	go func() {
		s.wgp.Wait()
		stopChan <- true
	}()
	select {
	case <-stopChan:
		return nil
	case <-time.After(time.Second * 10):
		return errors.New("SCC Close Timeout")
	}
}

//判断是否已经关闭
func (s *SCC) Done() bool {
	return Done(s.ctx)
}

func Done(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
