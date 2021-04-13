package utils

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

func NewSCC() *SCC {
	s := &SCC{
		stopCancel: make(chan struct{}),
	}
	s.wgp.Add(1)
	return s
}

//协程控制器
type SCC struct {
	wgp        sync.WaitGroup
	stopMutex  int32
	stopCancel chan struct{}
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
func (s *SCC) CGO(f func(chan struct{})) {
	go func() {
		s.wgp.Add(1)
		defer s.wgp.Done()
		f(s.stopCancel)
	}()
}

func (s *SCC) Wait() {
	s.wgp.Wait()
}

//Close
func (s *SCC) Close() error {
	if !atomic.CompareAndSwapInt32(&s.stopMutex, 0, 1) {
		return errors.New("scc Close exist")
	}
	s.wgp.Done()
	close(s.stopCancel)
	stopTimeout := make(chan bool)
	go func() {
		s.wgp.Wait()
		stopTimeout <- true
	}()
	select {
	case <-stopTimeout:
		return errors.New("SCC Close success")
	case <-time.After(time.Second * 10):
		return errors.New("SCC Close timeout")
	}

}

func (s *SCC) Stopped() bool {
	return s.stopMutex == 1
}
