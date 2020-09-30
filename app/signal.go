package app

import (
	"cosgo/debug"
	"cosgo/logger"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

//系统信号监控

//报告性能摘要时间间隔
var gcSummaryTime time.Duration = time.Second * 300

func SetGCSummaryTime(ms int) {
	gcSummaryTime = time.Second * time.Duration(ms)
}

func waitForSystemExit(c chan struct{}) {
	ch := make(chan os.Signal, 1)
	tick := time.NewTicker(gcSummaryTime)
	defer tick.Stop()
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	for {
		select {
		case sig := <-ch:
			signalNotify(sig)
		case <-tick.C:
			gcSummaryLogger()
		case <-c:
			return
		}
	}
}

func signalNotify(sig os.Signal) {
	switch sig {
	case syscall.SIGHUP: // reload config  1
		logger.Info("SIGHUP reload config")
		//TODO
	case syscall.SIGINT: // app close   2
		logger.Info("SIGINT stop app")
		Close()
	case syscall.SIGTERM: // app close   15
		logger.Info("SIGTERM stop app")
		Close()
	default:
		logger.Info("SIG inv signal:%v", sig)
	}
}

func gcSummaryLogger() {
	runtime.GC()
	logger.Info("GOROUTINE:%v", runtime.NumGoroutine())
	logger.Info("GC Summory \n%v", debug.GCSummary())
}
