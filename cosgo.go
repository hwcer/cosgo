package cosgo

import (
	"fmt"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"time"

	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/logger"
)

var modules []Module

//var appStopChan = make(chan struct{})

//func assert(err interface{}, s ...string) {
//	if err != nil {
//		logger.Sprint(logger.LevelFATAL, logger.Format(err), s[0])
//	} else if len(s) > 0 {
//		logger.Trace(s[0])
//	}
//}

func Use(mods ...Module) {
	for _, mod := range mods {
		modules = append(modules, mod)
	}
}
func Range(f func(Module) bool) {
	for _, mod := range modules {
		if !f(mod) {
			return
		}
	}
}

// Start 应用程序启动
// @param waitForSystemExit 阻塞模式等待系统关闭系统，程序不会直接退出
// @param mods 需注册的模块

func Start(waitForSystemExit bool, mods ...Module) {
	for _, mod := range mods {
		modules = append(modules, mod)
	}
	var err error
	if err = Config.Init(); err != nil {
		logger.Panic(err)
	}
	if err = helps(); err != nil {
		panic(err)
	}
	if err = writePidFile(); err != nil {
		logger.Panic(err)
	}
	emit(EventTypBegin)
	logger.Trace("App Starting")
	defer func() {
		if err = deletePidFile(); err != nil {
			logger.Trace("App delete pid file err:%v", err)
		}
		logger.Trace("App Closed\n")
		emit(EventTypStopped)
		_ = logger.Close()
	}()
	//=========================加载模块=============================
	if err = pprofStart(); err != nil {
		logger.Sprint(logger.LevelFatal, logger.Format(err), string(debug.Stack()))
	}
	defer func() {
		_ = pprofClose()
	}()
	//assert(emit(EventTypInitBefore))
	for _, v := range modules {
		if err = v.Init(); err != nil {
			logger.Sprint(logger.LevelFatal, logger.Format(err), string(debug.Stack()))
		} else {
			logger.Trace("mod[%v] init", v.Id())
		}
	}
	emit(EventTypLoaded)
	//自定义进程
	if Options.Process != nil && !Options.Process() {
		return
	}
	//=========================启动信息=============================
	showConfig()
	//=========================启动模块=============================
	for _, v := range modules {
		scc.Add(1)
		if err = v.Start(); err != nil {
			logger.Sprint(logger.LevelFatal, logger.Format(err), string(debug.Stack()))
		} else {
			logger.Trace("mod[%v] start", v.Id())
		}
	}
	emit(EventTypStarted)
	Options.Banner()
	if waitForSystemExit {
		WaitForSystemExit()
	}
}

// Close 外部关闭程序
func Close() bool {
	return stop()
}

func showConfig() {
	var log []string
	log = append(log, "\n============================Show App Config============================")
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
	emit(EventTypClosing)
	logger.Alert("App will stop")
	for i := len(modules) - 1; i >= 0; i-- {
		closeModule(modules[i])
	}
	if err := scc.Wait(time.Second * 10); err != nil {
		logger.Alert("App Stop Error:%v", err)
	}
	return true
}

func closeModule(m Module) {
	defer scc.Done()
	defer func() {
		if err := recover(); err != nil {
			logger.Sprint(logger.LevelError, logger.Format(err), string(debug.Stack()))
		}
	}()
	if err := m.Close(); err != nil {
		logger.Sprint(logger.LevelError, logger.Format(err), string(debug.Stack()))
	} else {
		logger.Trace("mod [%v] stop", m.Id())
	}
}
