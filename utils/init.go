package utils

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
)

var stop int32 //停止标志

var stopCancel context.CancelFunc
var stopContext context.Context
var stopWaitGroup sync.WaitGroup

var EventWriteChanSize = 5000
var WorkerWriteChanSize = 5000

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	stopWaitGroup.Add(1)
	stopContext, stopCancel = context.WithCancel(context.Background())
}

func Stop(wait ...bool) {
	if !atomic.CompareAndSwapInt32(&stop, 0, 1) {
		fmt.Printf("Server Stop error\n")
		return
	}
	stopCancel()
	stopWaitGroup.Done()
	if len(wait) > 0 && wait[0] {
		stopWaitGroup.Wait()
	}
}

func IsStop() bool {
	return stop == 1
}

func WaitForSystemExit() {
	var stopChanForSys = make(chan os.Signal, 1)
	signal.Notify(stopChanForSys, os.Interrupt, os.Kill, syscall.SIGTERM)
	select {
	case <-stopChanForSys:
		Stop()
	}
	close(stopChanForSys)
}

func Go(fn func(ctx context.Context)) {
	go func() {
		defer func() {
			stopWaitGroup.Done()
			if err := recover(); err != nil {
				fmt.Printf("panic in Go: %v\n", err)
			}
		}()
		stopWaitGroup.Add(1)
		fn(stopContext)
	}()
}
