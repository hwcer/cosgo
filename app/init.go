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

func init() {
	cancel = make(chan struct{})
}

func Go(fn func()) {
	wgp.Add(1)
	go func() {
		defer wgp.Done()
		fn()
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

func Timeout(d time.Duration, fn func() error) error {
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
