package app


import (
	"context"
	"cosgo/logger"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"
)

/*
使用event完成 模块的init
 */

var (
	wgp      sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	mainFun func()
	modules []Module
)

func init()  {
	ctx, cancel = context.WithCancel(context.Background())
	//wgp = &sync.WaitGroup{}
}

func assert(err error, s string) {
	if err != nil {
		logger.Fatal("app boot failed, %v: %v", s, err)
	} else {
		logger.Info("app boot %v done", s)
	}
}

func SetMain(m func()) {
	mainFun = m
}


/**
 * 应用程序启动
 * @param name 应用程序名
 * @param defConf 默认配置文件名
 * @param iFlag  自定义的flag配置，可以为空
 * @param m 需注册的模块
 */
func Start( m ...Module) {
	modules = m
	assert(appReady(), fmt.Sprintf("init main"))
	for _, v := range m {
		assert(v.Ready(), fmt.Sprintf("mod [%v] init", v.ID()))
	}
	assert(appStart(), fmt.Sprintf("main start"))
	for _, v := range m {
		assert(v.Start(ctx, &wgp), fmt.Sprintf("mod [%v] start", v.ID()))
	}
	// pid
	assert(initPidFile(), "init pidfile")
	if mainFun != nil {
		mainFun()
	}
	wgp.Wait()
	deletePidFile()
	logger.Warn("Say byebye to the world")
}

func Stop() error {
	logger.Info("App will stop")

	for _, v := range modules {
		func(m Module) {
			err := m.Stop()
			logger.Info("mod [%v] stop result:%v", m.ID(), err)
		}(v)
	}

	wgp.Done()
	logger.Info("App stop done")
	return nil
}

func Exit() {
	cancel()
}

//APP控制器 ready
func appReady() error  {
	// 随机种子
	rand.Seed(time.Now().UnixNano())
	// 命令行解析
	assert(initFlag(), "init flag")
	// 初始性能调优
	assert(initProfile(), "init profile")

	return nil
}
//APP控制器 start
func appStart() error  {
	wgp.Add(1)
	// 输出基本配置项
	showConfig()

	go appWorker()
	return nil
}


func showConfig() {
	logger.Info("=============== show app config ======================")
	logger.Info(">> AppName:%v", appName)
	logger.Info(">> AppBinDir:%v", appBinDir)
	logger.Info(">> AppExecDir:%v", appExecDir)
	logger.Info(">> AppWorkDir:%v", appWorkDir)
	logger.Info(">> AppLogDir:%v", Flag.GetString("logdir"))
	logger.Info(">> AppPidFile:%v", Flag.GetString("pidfile"))

	logger.Info(">> CPU:%v  Pid:%v", runtime.NumCPU(), os.Getpid())

	logger.Info("======================================================")
}

//系统信号
func appWorker()  {
	timeSec := time.Second * 30
	tickSec := time.NewTimer(timeSec)

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

	for {
		select {
		case sig := <-sigCh:
			signalNotify(sig)
		case <-ctx.Done():
			{
				Stop()
				return
			}
		case <-tickSec.C:
			{
				tickSec.Reset(timeSec)
				runtime.GC()
				logger.Info("NUM GOROUTINE:%v", runtime.NumGoroutine())
				//logger.Info("GC Summory \n%v", debug.GCSummary())
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
			cancel()
		case syscall.SIGTERM: // app close   15
			logger.Info("SIGTERM stop app")
			cancel()
		default:
			logger.Info("SIG inv signal:%v", sig)
	}
}

