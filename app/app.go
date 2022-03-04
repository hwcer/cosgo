package app

import (
	"fmt"
	"github.com/hwcer/cosgo/library/logger"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	banner         func()
	modules        []Module
	DataTimeFormat = "2006-01-02 15:04:05 -0700"
)

func assert(err interface{}, s ...string) {
	if err != nil {
		panic(fmt.Sprintf("%v failed: %v\n", s, err))
	} else if len(s) > 0 {
		fmt.Printf("app %v done\n", s[0])
	}
}

func defBanner() {
	str := `
  _________  _____________ 
 / ___/ __ \/ __/ ___/ __ \
/ /__/ /_/ /\ \/ (_ / /_/ /
\___/\____/___/\___/\____/
____________________________________O/_______
                                    O\
`
	fmt.Printf(str)
}

//设置默认启动界面，启动完所有MOD后执行
func SetBanner(m func()) {
	banner = m
}

func Use(mods ...Module) {
	for _, mod := range mods {
		modules = append(modules, mod)
	}
}

/**
 * 应用程序启动
 * @param m 需注册的模块
 */
func Start(mods ...Module) {
	fmt.Printf("App Starting:%v\n", time.Now().Format(DataTimeFormat))
	for _, mod := range mods {
		modules = append(modules, mod)
	}
	rand.Seed(time.Now().UnixNano())
	var err error
	if err = initFlag(); err != nil {
		panic(err)
	}
	if err = initBuild(); err != nil {
		panic(err)
	}
	if err = writePidFile(); err != nil {
		panic(err)
	}
	defer func() {
		if err = deletePidFile(); err != nil {
			fmt.Printf("App delete pid file err:%v\n", err)
		}
		fmt.Printf("App Closed:%v\n", time.Now().Format(DataTimeFormat))
	}()
	//=========================加载模块=============================
	if err = pprofStart(); err != nil {
		panic(err)
	}
	defer pprofClose()
	assert(emit(EventTypInitBefore))
	for _, v := range modules {
		assert(v.Init(), fmt.Sprintf("mod [%v] init", v.ID()))
	}
	assert(emit(EventTypInitAfter))
	//=========================启动信息=============================
	showConfig()
	//=========================启动模块=============================
	assert(emit(EventTypStartBefore))
	for _, v := range modules {
		SCC.WaitGroup.Add(1)
		assert(v.Start(), fmt.Sprintf("mod [%v] start", v.ID()))
	}
	assert(emit(EventTypStartAfter))
	if banner != nil {
		banner()
	} else {
		defBanner()
	}
	if loggerConsoleAdapter != nil {
		loggerConsoleAdapter.Options.Format = nil
	}
	WaitForSystemExit()
	//fmt.Printf("App Wait Done\n")
}

func Close() {
	if !SCC.Cancel() {
		return
	}
	fmt.Printf("App will stop\n")
	assert(emit(EventTypCloseBefore))
	for i := len(modules) - 1; i >= 0; i-- {
		closeModule(modules[i])
	}
	assert(emit(EventTypCloseAfter))
	if err := SCC.Wait(time.Second * 30); err != nil {
		fmt.Printf("App Stop Err:%v\n", err)
	}
}

func closeModule(m Module) {
	defer SCC.WaitGroup.Done()
	defer func() {
		if err := recover(); err != nil {
			logger.Error("%v", err)
		}
	}()
	assert(m.Close(), fmt.Sprintf("mod [%v] stop", m.ID()))
}

func showConfig() {
	var log []string
	log = append(log, "")
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

	log = append(log, fmt.Sprintf(">> BUIND GO:%v VER:%v  TIME:%v", BUIND_GO, BUIND_VER, BUIND_TIME))
	log = append(log, fmt.Sprintf(">> RUNTIME GO:%v  CPU:%v  Pid:%v", runtime.Version(), runtime.NumCPU(), os.Getpid()))
	log = append(log, "========================================================================")
	log = append(log, "")
	fmt.Printf(strings.Join(log, "\n"))
}
