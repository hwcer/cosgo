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
	banner  func()
	modules []Module
)

func assert(err interface{}, s string) {
	if err != nil {
		logger.Fatal("app failed, %v: %v", s, err)
	} else {
		logger.Info("app %v done", s)
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
	logger.Info(str)
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
		deletePidFile()
		fmt.Printf("Say byebye to the world")
	}()
	//=========================加载模块=============================
	initProfile()

	for _, v := range modules {
		assert(v.Init(), fmt.Sprintf("mod [%v] init", v.ID()))
	}
	//=========================启动信息=============================
	showConfig()
	//=========================启动模块=============================
	for _, v := range modules {
		wgp.Add(1)
		assert(v.Start(), fmt.Sprintf("mod [%v] start", v.ID()))
	}

	if banner != nil {
		banner()
	} else {
		defBanner()
	}

	wgp.Add(1)
	Go(waitForSystemExit)
	wgp.Wait()
}

func Close() {
	logger.Info("App will stop")
	if !atomic.CompareAndSwapInt32(&stop, 0, 1) {
		logger.Error("Server Close error")
		return
	}
	defer wgp.Done()

	for _, v := range modules {
		func(m Module) {
			defer func() {
				wgp.Done()
				if err := recover(); err != nil {
					logger.Error("%v", err)
				}
			}()
			assert(v.Close(), fmt.Sprintf("mod [%v] stop", m.ID()))
		}(v)
	}
	if cancel != nil {
		close(cancel)
	}
	logger.Info("App stop done")
}

func showConfig() {
	var log []string
	log = append(log, "")
	log = append(log, "====================== Show App Config ======================")
	log = append(log, fmt.Sprintf(">> AppName:%v", GetName()))
	log = append(log, fmt.Sprintf(">> AppRoot:%v", GetDir()))
	log = append(log, fmt.Sprintf(">> AppLogs:%v", Config.GetString("logs")))
	log = append(log, fmt.Sprintf(">> PidFile:%v", Config.GetString("pidfile")))
	log = append(log, fmt.Sprintf(">> BUIND GO:%v VER:%v  TIME:%v", BUIND_GO, BUIND_VER, BUIND_TIME))
	log = append(log, fmt.Sprintf(">> RUNTIME GO:%v  CPU:%v  Pid:%v", runtime.Version(), runtime.NumCPU(), os.Getpid()))
	log = append(log, "=============================================================")
	log = append(log, "")
	logger.Info(strings.Join(log, "\n"))
}
