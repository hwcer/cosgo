package app

import (
	"fmt"
	"github.com/hwcer/cosgo/logger"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"time"
)

var (
	banner  func()
	modules []Module
)

func assert(err interface{}, s string) {
	if err != nil {
		fmt.Printf("app failed, %v: %v\n", s, err)
	} else {
		fmt.Printf("app %v done\n", s)
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
	fmt.Printf("App Starting:%v\n", time.Now().String())
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
		pprofClose()
		deletePidFile()
		fmt.Printf("App Closed:%v\n", time.Now().String())
	}()
	//=========================加载模块=============================
	if err = pprofStart(); err != nil {
		panic(err)
	}

	for _, v := range modules {
		assert(v.Init(), fmt.Sprintf("mod [%v] init", v.ID()))
	}
	//=========================启动信息=============================
	showConfig()
	//=========================启动模块=============================
	for _, v := range modules {
		scc.Add(1)
		assert(v.Start(), fmt.Sprintf("mod [%v] start", v.ID()))
	}

	if banner != nil {
		banner()
	} else {
		defBanner()
	}

	WaitForSystemExit()
	//fmt.Printf("App Wait Done\n")
}

func Close() {
	if !scc.Close() {
		return
	}
	fmt.Printf("App will stop\n")
	for _, m := range modules {
		closeModule(m)
	}
	if err := scc.Wait(time.Second * 30); err != nil {
		fmt.Printf("App Stop Err:%v\n", err)
	}
}

func closeModule(m Module) {
	defer func() {
		scc.Done()
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
	log = append(log, fmt.Sprintf(">> AppName:%v", GetName()))
	log = append(log, fmt.Sprintf(">> AppRoot:%v", GetDir()))
	log = append(log, fmt.Sprintf(">> AppLogs:%v", Config.GetString("logs")))
	log = append(log, fmt.Sprintf(">> PidFile:%v", Config.GetString("pidfile")))
	log = append(log, fmt.Sprintf(">> BUIND GO:%v VER:%v  TIME:%v", BUIND_GO, BUIND_VER, BUIND_TIME))
	log = append(log, fmt.Sprintf(">> RUNTIME GO:%v  CPU:%v  Pid:%v", runtime.Version(), runtime.NumCPU(), os.Getpid()))
	log = append(log, "========================================================")
	log = append(log, "")
	fmt.Printf(strings.Join(log, "\n"))
}
