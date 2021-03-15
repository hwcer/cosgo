package app

import (
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

func waitForSystemExit() {
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
		case <-cancel:
			return
		}
	}
}

func signalNotify(sig os.Signal) {
	logger.Debug("OS SIGINT:%v", sig)
	switch sig {
	case syscall.SIGHUP: // reload Config  1
		logger.Info("SIGHUP reload Config")
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
	//logger.Info("GC Summory \n%v", debug.GCSummary())
}
