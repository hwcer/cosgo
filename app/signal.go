package app


import (
	"context"
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
var gcSummaryTime time.Duration

func init()  {
	gcSummaryTime = time.Second * 300
}

func SetGCSummaryTime(ms int)  {
	gcSummaryTime = time.Second * time.Duration(ms)
}



func waitForSystemExit(c context.Context)  {
	ch := make(chan os.Signal, 1)
	tick := time.NewTicker(gcSummaryTime)
	signal.Notify(ch, syscall.SIGHUP, syscall.SIGQUIT,syscall.SIGINT,syscall.SIGKILL)

	for {
		select {
		case sig := <-ch:
			signalNotify(sig)
		case <-c.Done():
			return
		case <-tick.C:
			{
				runtime.GC()
				logger.Info("GOROUTINE:%v", runtime.NumGoroutine())
				logger.Info("GC Summory \n%v", debug.GCSummary())
			}
		}
	}
}



func signalNotify(sig os.Signal)  {
	switch sig {
		case syscall.SIGHUP: // reload config  1
			logger.Info("SIGHUP reload config")
			//TODO
		case syscall.SIGINT: // app close   2
			logger.Info("SIGINT stop app")
			Stop()
		case syscall.SIGTERM: // app close   15
			logger.Info("SIGTERM stop app")
			Stop()
		default:
			logger.Info("SIG inv signal:%v", sig)
	}
}

