package main

import (
	"context"
	"github.com/hwcer/cosgo"
	"github.com/hwcer/cosgo/logger"
	"time"
)

func main() {
	cosgo.Start(&module{Module: cosgo.NewModule("test")})
	cosgo.WaitForSystemExit()
}

func timer(ctx context.Context) {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			logger.Info("timer")
		}
	}
}

type module struct {
	*cosgo.Module
}

func (this *module) Init() error {
	logger.Debug("module Init")
	cosgo.SCC.CGO(timer)
	return nil
}

func (this *module) Start() error {
	logger.Debug("module Start")
	return nil
}

func (this *module) Close() error {
	logger.Debug("module Close")
	return nil
}
