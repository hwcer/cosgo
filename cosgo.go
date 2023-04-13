package cosgo

import (
	"fmt"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/logger"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"
)

var modules []IModule

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

func Modules() (r []string) {
	for _, m := range modules {
		r = append(r, m.ID())
	}
	return
}

/**
 * 应用程序启动
 * @param mods 需注册的模块
 */
func Start(waitForSystemExit bool, mods ...IModule) {
	for _, mod := range mods {
		modules = append(modules, mod)
	}
	rand.Seed(time.Now().UnixNano())
	var err error
	if err = Config.init(); err != nil {
		panic(err)
	}
	if err = helps(); err != nil {
		panic(err)
	}
	if err = writePidFile(); err != nil {
		panic(err)
	}
	logger.Trace("\n")
	logger.Trace("App Starting")
	defer func() {
		if err = deletePidFile(); err != nil {
			logger.Trace("App delete pid file err:%v", err)
		}
		logger.Trace("App Closed")
	}()
	//=========================加载模块=============================
	if err = pprofStart(); err != nil {
		panic(err)
	}
	defer func() {
		_ = pprofClose()
	}()
	assert(emit(EventTypInitBefore))
	for _, v := range modules {
		assert(v.Init(), fmt.Sprintf("mod[%v] init", v.ID()))
	}
	assert(emit(EventTypInitAfter))
	//自定义进程
	if Options.Process != nil && !Options.Process() {
		return
	}
	//=========================启动信息=============================
	showConfig()
	//=========================启动模块=============================
	assert(emit(EventTypStartBefore))
	for _, v := range modules {
		scc.Add(1)
		assert(v.Start(), fmt.Sprintf("mod[%v] start", v.ID()))
	}
	assert(emit(EventTypStartAfter))
	Options.Banner()
	if waitForSystemExit {
		WaitForSystemExit()
	}
}

func Close() {
	if !scc.Cancel() {
		return
	}
	logger.Alert("App will stop")
	assert(emit(EventTypCloseBefore))
	for i := len(modules) - 1; i >= 0; i-- {
		closeModule(modules[i])
	}
	assert(emit(EventTypCloseAfter))
	if err := scc.Wait(time.Second * 30); err != nil {
		logger.Error("App Stop Error:%v", err)
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
	log = append(log, fmt.Sprintf(">> appName:%v", Name()))
	log = append(log, fmt.Sprintf(">> workDir:%v", WorkDir()))

	logsDir := Config.GetString(AppConfigNameLogsDir)
	if logsDir == "" {
		logsDir = "Console"
	}
	log = append(log, fmt.Sprintf(">> logsDir:%v", logsDir))

	pidfile := ""
	if enablePidFile {
		pidfile = Config.GetString(AppConfigNamePidFile)
	} else {
		pidfile = "Disable"
	}
	log = append(log, fmt.Sprintf(">> pidFile:%v", pidfile))

	//log = append(log, fmt.Sprintf(">> BUIND GO:%v VER:%v  TIME:%v", BUIND_GO, BUIND_VER, BUIND_TIME))
	log = append(log, fmt.Sprintf(">> RUNTIME GO:%v  CPU:%v  Pid:%v", runtime.Version(), runtime.NumCPU(), os.Getpid()))
	log = append(log, "========================================================================")
	log = append(log, "")
	logger.Trace(strings.Join(log, "\n"))
}
