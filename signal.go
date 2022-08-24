package cosgo

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"
)

//系统信号监控

// 报告性能摘要时间间隔
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
		case <-SCC.Context.Done():
			return
		}
	}
}

func signalNotify(sig os.Signal) {
	switch sig {
	case syscall.SIGHUP: // reload Config  1
		fmt.Printf("SIGHUP reload Config\n")
	case syscall.SIGINT, syscall.SIGTERM: // app close   2
		fmt.Printf("SIGINT Stop App:%v\n", sig)
		Close()
	default:
		fmt.Printf("SIG inv signal:%v\n", sig)
	}
}

func gcSummaryLogger() {
	runtime.GC()
	fmt.Printf("GOROUTINE:%v\n", runtime.NumGoroutine())
	//logger.Info("GC Summory \n%v", debug.GCSummary())
}
