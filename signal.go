package cosgo

import (
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/logger"
)

// SignalReload 重新加载系统信号 kill -10 pid
var SignalReload syscall.Signal = 0xa

// GCSummaryTime 报告性能摘要时间间隔
var GCSummaryTime time.Duration = time.Second * 300

func WaitForSystemExit() {
	ch := make(chan os.Signal, 1)
	timer := time.NewTimer(GCSummaryTime)
	defer stop()
	defer timer.Stop()
	ctx, cancel := scc.WithCancel()
	defer cancel()
	signal.Notify(ch, syscall.SIGINT, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM, SignalReload)
	for {
		select {
		case sig := <-ch:
			if stopped := signalNotify(sig); stopped {
				return
			}
		case <-timer.C:
			gcSummaryLogs()
			timer.Reset(GCSummaryTime)
		case <-ctx.Done():
			return
		}
	}
}

// 系统信号监控
// syscall.SIGINT 控制台按下 CTRL+C
// syscall.SIGHUP 关闭控制台，无论是否以服务模式启动，程序都会收到这个信号，使用nohup 启动避免程序退出
// syscall.SIGQUIT 退出
// syscall.SIGKILL  kill -9 系统强制退出程序
// syscall.SIGTERM  kill 无参数时默认信号
func signalNotify(sig os.Signal) (stopped bool) {
	logger.Trace("收到信号：%v\n", sig)
	switch sig {
	case SignalReload:
		Reload()
	case syscall.SIGHUP:
		SIGHUP()
	case syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM:
		stopped = true
	default:
		logger.Trace("receive signal:%v", sig)
	}
	return
}

// SIGHUP 关闭控制台
func SIGHUP() {
	logger.Trace("停止控制台输出")
	logger.Console.Disable = true
}

func gcSummaryLogs() {
	runtime.GC()
	logger.Info("GOROUTINE:%v", runtime.NumGoroutine())
}
