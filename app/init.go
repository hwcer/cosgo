package app

import (
	"cosgo/logger"
	"fmt"
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
	go func() {
		defer func() {
			wgp.Done()
			if err := recover(); err != nil {
				fmt.Printf("panic in Go: %v\n", err)
			}
		}()
		wgp.Add(1)
		fn()
	}()
}

//带cancel的GO协程
func Go2(fn func(chan struct{})) {
	go func() {
		defer func() {
			wgp.Done()
			if err := recover(); err != nil {
				fmt.Printf("panic in Go: %v\n", err)
			}
		}()
		wgp.Add(1)
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
