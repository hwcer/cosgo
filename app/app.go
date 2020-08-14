package app

import (
	"cosgo/logger"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"sync/atomic"
	"time"
)


var (
	mainFun func()
	modules []Module
)



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
	assert(appInit(), fmt.Sprintf("init main"))
	for _, v := range m {
		assert(v.Init(), fmt.Sprintf("mod [%v] init", v.ID()))
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

func Stop() {
	logger.Info("App will stop")
	if !atomic.CompareAndSwapInt32(&stop, 0, 1) {
		logger.Error("Server Stop error")
		return
	}

	for _, v := range modules {
		func(m Module) {
			err := m.Stop()
			logger.Info("mod [%v] stop result:%v", m.ID(), err)
		}(v)
	}
	cancel()
	wgp.Done()
	logger.Info("App stop done")
}


//APP控制器 ready
func appInit() error  {
	// 随机种子
	rand.Seed(time.Now().UnixNano())
	// 命令行解析
	assert(initFlag(), "init flag")
	assert(buildInit(), "init build args")
	// 初始性能调优
	assert(initProfile(), "init profile")

	return nil
}
//APP控制器 start
func appStart() error  {
	wgp.Add(1)
	// 输出基本配置项
	showConfig()

	Go(waitForSystemExit)
	return nil
}


func showConfig() {
	logger.Info("=============== show app config ======================")
	logger.Info(">> AppName:%v", Flag.GetString("name"))
	logger.Info(">> AppBinDir:%v", Flag.GetString("AppBinDir"))
	logger.Info(">> AppWorkDir:%v", Flag.GetString("AppWorkDir"))
	logger.Info(">> appExecFile:%v", Flag.GetString("appExecFile"))
	logger.Info(">> AppLogDir:%v", Flag.GetString("logdir"))
	logger.Info(">> AppPidFile:%v", Flag.GetString("pidfile"))

	logger.Info(">> CPU:%v  Pid:%v", runtime.NumCPU(), os.Getpid())

	logger.Info("======================================================")
}
