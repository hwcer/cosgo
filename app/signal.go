package app

import (
	"github.com/hwcer/cosgo/logger"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

//系统信号监控

//报告性能摘要时间间隔
var gcSummaryTime time.Duration = time.Second * 300

func SetGCSummaryTime(second int) {
	gcSummaryTime = time.Second * time.Duration(second)
}

func WaitForSystemExit() {
	ch := make(chan os.Signal, 1)
	timer := time.NewTimer(gcSummaryTime)
	defer timer.Stop()
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM)
	for {
		select {
		case sig := <-ch:
			signalNotify(sig)
		case <-timer.C:
			timer.Reset(gcSummaryTime)
			gcSummaryLogger()
		case <-scc.Context().Done():
			return
		}
	}
}

func signalNotify(sig os.Signal) {
	logger.Debug("OS SIGINT:%v", sig)
	switch sig {
	case syscall.SIGHUP: // reload Config  1
		logger.Info("SIGHUP reload Config")
	case syscall.SIGINT, syscall.SIGTERM: // app close   2
		logger.Info("SIGINT stop app")
		Close()
	default:
		logger.Info("SIG inv signal:%v", sig)
	}
}

func gcSummaryLogger() {
	runtime.GC()
	logger.Info("GOROUTINE:%v", runtime.NumGoroutine())
	//logger.Info("GC Summory \n%v", debug.GCSummary())
}
