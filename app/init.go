package app

import (
	"cosgo/logger"
	"sync"
	"time"
)

var (
	wgp    sync.WaitGroup
	stop   int32 //停止标志
	cancel chan struct{}
)

var Debug bool

func init() {
	cancel = make(chan struct{})
}

type Module interface {
	ID() string
	Load() error
	Start(*sync.WaitGroup) error
	Close(*sync.WaitGroup) error
}

func Go(fn func()) {
	wgp.Add(1)
	go func() {
		defer func() {
			wgp.Done()
			if err := recover(); err != nil {
				logger.Error("panic in Go: %v\n", err)
			}
		}()
		fn()
	}()
}

//带cancel的GO协程
func Go2(fn func(chan struct{})) {
	go func() {
		wgp.Add(1)
		defer func() {
			wgp.Done()
			if err := recover(); err != nil {
				logger.Error("panic in Go: %v\n", err)
			}
		}()
		fn(cancel)
	}()
}

func Try(fun func(), handler ...func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			if len(handler) == 0 {
				logger.Error("%v", err)
			} else {
				handler[0](err)
			}
		}
	}()
	fun()
}

func TimeOut(d time.Duration, fn func() error) error {
	cher := make(chan error)

	go func() {
		cher <- fn()
	}()

	var err error
	select {
	case e := <-cher:
		err = e
	case <-time.After(d):
		err = nil
	}

	return err
}
