package utils

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
)

var EventWriteChanSize = 5000
var WorkerWriteChanSize = 5000

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func Try(f func(), handler ...func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			if len(handler) == 0 {
				fmt.Printf("%v", err)
			} else {
				handler[0](err)
			}
		}
	}()
	f()
}

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
		return errors.New("scc Close success")
	case <-time.After(time.Second * 10):
		return errors.New("scc Close timeout")
	}

}

func (s *SCC) Stopped() bool {
	return s.stopMutex == 1
}
