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
	s.wgp = new(sync.WaitGroup)
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.wgp.Add(1)
	return s
}

//协程控制器
type SCC struct {
	wgp    *sync.WaitGroup
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

func (s *SCC) Wait() {
	s.wgp.Wait()
}

func (s *SCC) Context() context.Context {
	return s.ctx
}

func (s *SCC) WaitGroup() *sync.WaitGroup {
	return s.wgp
}

//Close ,callback:成功调用Close后 cancel之前调用
func (s *SCC) Close(cb ...func()) error {
	if !atomic.CompareAndSwapInt32(&s.stop, 0, 1) {
		return errors.New("SCC Close Exist")
	}
	if len(cb) > 0 {
		cb[0]()
	}
	s.cancel()
	s.wgp.Done()
	//fmt.Printf("SCC CLOSE\n")
	return Timeout(time.Second*10, func() error {
		s.wgp.Wait()
		//fmt.Printf("SCC wgp.Wait Done\n")
		return nil
	})
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
