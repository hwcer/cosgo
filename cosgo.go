package cosgo

import (
	"fmt"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/logger"
	"os"
	"runtime"
	"strings"
	"time"
)

var modules []IModule
var appStopChan = make(chan struct{})

func assert(err interface{}, s ...string) {
	if err != nil {
		logger.Fatal(err)
	} else if len(s) > 0 {
		logger.Trace(s[0])
	}
}

func Use(mods ...IModule) {
	for _, mod := range mods {
		modules = append(modules, mod)
	}
}
func Range(f func(IModule) bool) {
	for _, mod := range modules {
		if !f(mod) {
			return
		}
	}
}

/**
 * 应用程序启动
 * @param mods 需注册的模块
 */
func Start(waitForSystemExit bool, mods ...IModule) {
	for _, mod := range mods {
		modules = append(modules, mod)
	}
	var err error
	if err = Config.init(); err != nil {
		logger.Panic(err)
	}
	if err = helps(); err != nil {
		panic(err)
	}
	if err = writePidFile(); err != nil {
		logger.Panic(err)
	}
	assert(emit(EventTypStarting))
	logger.Trace("App Starting")
	defer func() {
		if err = deletePidFile(); err != nil {
			logger.Trace("App delete pid file err:%v", err)
		}
		logger.Trace("App Closed\n")
		assert(emit(EventTypStopped))
	}()
	//=========================加载模块=============================
	if err = pprofStart(); err != nil {
		logger.Panic(err)
	}
	defer func() {
		_ = pprofClose()
	}()
	//assert(emit(EventTypInitBefore))
	for _, v := range modules {
		assert(v.Init(), fmt.Sprintf("mod[%v] init", v.ID()))
	}
	//assert(emit(EventTypInitialize))
	//自定义进程
	if Options.Process != nil && !Options.Process() {
		return
	}
	//=========================启动信息=============================
	showConfig()
	//=========================启动模块=============================
	for _, v := range modules {
		scc.Add(1)
		assert(v.Start(), fmt.Sprintf("mod[%v] start", v.ID()))
	}
	Options.Banner()
	assert(emit(EventTypStarted))

	if waitForSystemExit {
		WaitForSystemExit()
	}
}

// Close 外部关闭程序
func Close(force bool) {
	if !(stop() || force) {
		return
	}
	select {
	case <-appStopChan:
	default:
		close(appStopChan)
	}
}

func closeModule(m IModule) {
	defer scc.Done()
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
	}()
	assert(m.Close(), fmt.Sprintf("mod [%v] stop", m.ID()))
}

func showConfig() {
	var log []string
	log = append(log, "============================Show App Config============================")
	log = append(log, fmt.Sprintf(">> App : %v", Name()))
	pidFile := ""
	if enablePidFile {
		pidFile = Config.GetString(AppConfigNamePidFile)
	} else {
		pidFile = "Disable"
	}
	log = append(log, fmt.Sprintf(">> Pid : %v", pidFile))

	log = append(log, fmt.Sprintf(">> Path : %v", Dir()))
	logsDir := Config.GetString(AppConfigNameLogsPath)
	if logsDir == "" {
		logsDir = "Console"
	}
	log = append(log, fmt.Sprintf(">> Logs : %v", logsDir))
	log = append(log, fmt.Sprintf(">> Version : %v", Version))
	log = append(log, fmt.Sprintf(">> Runtime GO:%v  CPU:%v  Pid:%v", runtime.Version(), runtime.NumCPU(), os.Getpid()))
	log = append(log, "========================================================================")
	logger.Trace(strings.Join(log, "\n"))
}

func stop() (stopped bool) {
	if !scc.Cancel() {
		return true
	}
	assert(emit(EventTypStopping))
	logger.Alert("App will stop")
	for i := len(modules) - 1; i >= 0; i-- {
		closeModule(modules[i])
	}
	if err := scc.Wait(time.Second * 10); err != nil {
		logger.Alert("App Stop Error:%v", err)
		return false
	}
	return true
}
