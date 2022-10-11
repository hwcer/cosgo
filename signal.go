package cosgo

import (
	"github.com/hwcer/logger"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

// ReloadSignal 重新加载系统信号 kill -10 pid
var ReloadSignal syscall.Signal = 10

// gcSummaryTime 报告性能摘要时间间隔
var gcSummaryTime time.Duration = time.Second * 300

func SetGCSummaryTime(second int) {
	gcSummaryTime = time.Second * time.Duration(second)
}

func WaitForSystemExit() {
	ch := make(chan os.Signal, 1)
	timer := time.NewTimer(gcSummaryTime)
	defer timer.Stop()
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM, ReloadSignal)
	for {
		select {
		case sig := <-ch:
			signalNotify(sig)
		case <-timer.C:
			timer.Reset(gcSummaryTime)
			gcSummaryLogger()
		case <-SCC.Context.Done():
			return
		}
	}
}

// 系统信号监控
func signalNotify(sig os.Signal) {
	switch sig {
	case ReloadSignal:
		Reload()
	case syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTERM:
		logger.Info("SIGINT Stop App:%v\n", sig)
		Close()
	default:
		logger.Info("SIG inv signal:%v\n", sig)
	}
}

func gcSummaryLogger() {
	runtime.GC()
	logger.Info("GOROUTINE:%v\n", runtime.NumGoroutine())
}
