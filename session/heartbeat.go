package session

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/logger"
)

// Heartbeat 全局心跳管理器，只需启动一次
var Heartbeat = heartbeat{}

type heartbeat struct {
	started int32
}

// Start 启动心跳守护协程（幂等，多次调用安全）
func (this *heartbeat) Start() {
	if !atomic.CompareAndSwapInt32(&this.started, 0, 1) {
		return
	}
	scc.CGO(this.daemon)
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
			this.tick()
			ticker.Reset(ts)
		}
	}
}

func (this *heartbeat) tick() {
	defer func() {
		if err := recover(); err != nil {
			logger.Alert("session heartbeat tick error:%v", err)
		}
	}()
	Emit(EventHeartbeat, Options.Heartbeat)
}
