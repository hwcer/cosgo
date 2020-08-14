package app

import (
	"context"
	"fmt"
	"sync"
)

var (
	wgp      sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	stop int32 //停止标志
)

func init() {
	ctx, cancel = context.WithCancel(context.Background())
}



func Go(fn func(context.Context)) {
	go func() {
		defer func() {
			wgp.Done()
			if err := recover(); err != nil {
				fmt.Printf("panic in Go: %v\n", err)
			}
		}()
		wgp.Add(1)
		fn(ctx)
	}()
}


type Module interface {
	ID() string
	Init() error
	Start(context.Context, *sync.WaitGroup) error
	Stop() error
}