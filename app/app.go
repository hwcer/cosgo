package app

import (
	"cosgo/logger"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync/atomic"
	"time"
)

var (
	appMain func()
	modules []Module
)

func assert(err error, s string) {
	if err != nil {
		logger.Fatal("app failed, %v: %v", s, err)
	} else {
		logger.Info("app %v done", s)
	}
}

func defAppMain() {
	banner := `
  _________  _____________ 
 / ___/ __ \/ __/ ___/ __ \
/ /__/ /_/ /\ \/ (_ / /_/ /
\___/\____/___/\___/\____/
____________________________________O/_______
                                    O\
`
	logger.Info(banner)
}

//设置默认启动界面，启动完所有MOD后执行
func SetAppMain(m func()) {
	appMain = m
}

func Use(mods ...Module) {
	for _, mod := range mods {
		modules = append(modules, mod)
	}
}

/**
 * 应用程序启动
 * @param name 应用程序名
 * @param defConf 默认配置文件名
 * @param iFlag  自定义的flag配置，可以为空
 * @param m 需注册的模块
 */
func Start(mods ...Module) {
	for _, mod := range mods {
		modules = append(modules, mod)
	}
	rand.Seed(time.Now().UnixNano())
	//=========================加载模块=============================
	initFlag()
	initBuild()
	initProfile()

	for _, v := range modules {
		assert(v.Init(), fmt.Sprintf("mod [%v] init", v.ID()))
	}
	//=========================启动信息=============================
	showConfig()
	//=========================启动模块=============================
	for _, v := range modules {
		assert(v.Start(wgp), fmt.Sprintf("mod [%v] start", v.ID()))
	}

	if appMain != nil {
		appMain()
	} else {
		defAppMain()
	}

	writePidFile()
	wgp.Add(1)
	Go(waitForSystemExit)
	wgp.Wait()
	deletePidFile()
	logger.Warn("Say byebye to the world")
}

func Close() {
	logger.Info("App will stop")
	if !atomic.CompareAndSwapInt32(&stop, 0, 1) {
		logger.Error("Server Close error")
		return
	}
	for _, v := range modules {
		func(m Module) {
			assert(m.Close(wgp), fmt.Sprintf("mod [%v] stop", m.ID()))
		}(v)
	}
	if cancel != nil {
		close(cancel)
	}
	wgp.Done()
	logger.Info("App stop done")
}

func showConfig() {
	var log []string
	log = append(log, "")
	log = append(log, "=============== show app config ======================")
	log = append(log, fmt.Sprintf(">> AppName:%v", Config.GetString("name")))
	log = append(log, fmt.Sprintf(">> AppBinDir:%v", Config.GetString("AppBinDir")))
	log = append(log, fmt.Sprintf(">> AppLogDir:%v", Config.GetString("logdir")))
	log = append(log, fmt.Sprintf(">> AppWorkDir:%v", Config.GetString("AppWorkDir")))
	log = append(log, fmt.Sprintf(">> appExecFile:%v", Config.GetString("appExecFile")))
	log = append(log, fmt.Sprintf(">> AppPidFile:%v", Config.GetString("pidfile")))
	log = append(log, fmt.Sprintf(">> BUIND GO:%v VER:%v  TIME:%v", BUIND_GO, BUIND_VER, BUIND_TIME))
	log = append(log, fmt.Sprintf(">> RUNTIME GO:%v  CPU:%v  Pid:%v", runtime.Version(), runtime.NumCPU(), os.Getpid()))
	log = append(log, "======================================================")
	log = append(log, "")
	logger.Info(strings.Join(log, "\n"))
}
