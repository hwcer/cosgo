package session

import (
	"context"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/logger"
	"sync/atomic"
	"time"
)

var Heartbeat = heartbeat{}

type heartbeat struct {
	started   int32
	listeners []func(int32)
}

func (this *heartbeat) Start() {
	if !atomic.CompareAndSwapInt32(&this.started, 0, 1) {
		return
	}
	scc.CGO(this.daemon)
}

func (this *heartbeat) On(f func(int32)) {
	this.listeners = append(this.listeners, f)
}

func (this *heartbeat) emit(i int32) {
	for _, l := range this.listeners {
		l(i)
	}
}

func (this *heartbeat) daemon(ctx context.Context) {
	if Options.Heartbeat == 0 {
		logger.Debug("session heartbeat not start,Because Options.Heartbeat is 0")
		return
	}

	ts := time.Second * time.Duration(Options.Heartbeat)
	ticker := time.NewTimer(ts)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			this.heartbeat()
			ticker.Reset(ts)
		}
	}
}

func (this *heartbeat) heartbeat() {
	defer func() {
		if err := recover(); err != nil {
			logger.Alert("session.memory daemon ticker error:%v", err)
		}
	}()
	this.emit(Options.Heartbeat)
}
