package cosgo

import (
	"fmt"
	"os"
	"runtime"
	"slices"
	"strings"

	"github.com/hwcer/cosgo/phase"
	"github.com/hwcer/cosgo/scc"
	"github.com/hwcer/logger"
)

var modules []Module

// Use 注册模块。🔴 仅限 cosgo.Start 之前调用:Starting 期间 append 会与 Start 的
// 模块遍历竞争,这里按 Init 级判定(比 Sealed 更严),封板前后调用只 Alert 提示并忽略
func Use(mods ...Module) {
	if phase.Get() != phase.Init {
		phase.Alert("cosgo.Use(%v)", len(mods))
		return
	}
	modules = append(modules, mods...)
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
	modules = append(modules, mods...)
	var err error
	if err = Config.Init(); err != nil {
		logger.Fatal("Config.Init err:%v", err)
	}
	if err = helps(); err != nil {
		logger.Fatal("helps err:%v", err)
	}
	if err = writePidFile(); err != nil {
		logger.Fatal("writePidFile err:%v", err)
	}

	logger.Info("App Starting")
	phase.Set(phase.Starting) //启动阶段时钟:后续 phase.Sealed() 守卫以此为基准
	defer func() {
		if err = deletePidFile(); err != nil {
			logger.Warn("App delete pid file err:%v", err)
		}
		logger.Info("App Closed")
		_ = emit(EventTypStopped, false)
		_ = logger.Close()
	}()

	if err = emit(EventTypBegin, true); err != nil {
		logger.Fatal("App Start error:%v", err)
		return
	}
	//=========================加载模块=============================
	if err = pprofStart(); err != nil {
		logger.Fatal("pprofStart err:%v", err)
	}
	defer func() {
		_ = pprofClose()
	}()
	for _, v := range modules {
		if err = v.Init(); err != nil {
			logger.Fatal("mod[%v] init err:%v", v.Id(), err)
		} else {
			logger.Info("mod[%v] init", v.Id())
		}
	}
	if err = emit(EventTypLoaded, true); err != nil {
		logger.Fatal("App Start error:%v", err)
		return
	}
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
			logger.Fatal("mod[%v] start err:%v", v.Id(), err)
		} else {
			logger.Info("mod[%v] start", v.Id())
		}
	}
	if err = emit(EventTypStarted, true); err != nil {
		logger.Fatal("App Start error:%v", err)
		return
	}

	// sealEvents 启动完成后封板(由 Cosgo.Start 在 EventTypStarted 发完后调用)。
	// 封板时钟统一由 phase 包承载,session 等其他需要封板的模块直接读 phase.Sealed(),
	// 根包不再级联调用各自的 Seal* 函数
	//启动完成,推进 phase 至 Started:封板时钟统一由 phase 承载,
	phase.Set(phase.Started)
	//事件表根/session、业务自定义的"仅启动期注册"守卫都直接读 phase.Sealed(),
	//不再需要根包逐一级联调用各自的 Seal* 函数(见 phase 包)
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
	log = append(log, "Show App Config\n========================================================================")
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
	logger.Info(strings.Join(log, "\n"))
}

func stop() (stopped bool) {
	if !scc.Cancel() {
		return true
	}
	phase.Set(phase.Closing)
	_ = emit(EventTypClosing, false)
	logger.Info("App will stop")
	for _, module := range slices.Backward(modules) {
		closeModule(module)
	}
	if err := scc.Wait(0); err != nil {
		logger.Warn("App Stop Error:%v", err)
	}
	phase.Set(phase.Closed)
	return true
}

func closeModule(m Module) {
	defer scc.Done()
	defer func() {
		if err := recover(); err != nil {
			logger.Info("mod[%v] close err:%v", m.Id(), err)
		}
	}()
	if err := m.Close(); err != nil {
		logger.Info("mod[%v] close err:%v", m.Id(), err)
	} else {
		logger.Info("mod [%v] stopped", m.Id())
	}
}
