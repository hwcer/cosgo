package app

import (
	"context"
	confer "cosgo/config"
	"fmt"
	"github.com/google/uuid"
	"log"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	goarg "github.com/alexflint/go-arg"
)

type appCtrl struct {
	Modules       []ModuleItf   // 所有模块
	Flag          FlagItf       // flag
	Conf          confer.Confer // 配置读取器
	BaseConf      *AppBaseConf  // app配置
	ConfName      string        // 配置文件名
	Main          func()
	ServerInitErr func(err error)

	wgp      *sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
	pidFile  string // pid文件名
	assLog   bool   // 是否允许输出到标准输出
	loadFunc func(res interface{}) (confer.Confer, error)

	rpcSrv *srpc.SRpcServer // app rpc服务
}

var checkStop = func(s context.CancelFunc, pid int) {
}

type baseFlagImpl struct {
	AppBaseFlag
}

func (b *baseFlagImpl) GetBase() *AppBaseFlag {
	return &b.AppBaseFlag
}

type modAppCmd struct {
	srpc.BaseRpcService
	ctrl *appCtrl
}

func (a *appCtrl) Run(defConf string, iFlag FlagItf, m ...ModuleItf) {
	a.ConfName = defConf
	a.Flag = iFlag
	a.Conf = nil
	a.Modules = m
	a.BaseConf = new(AppBaseConf)

	a.ctx, a.cancel = context.WithCancel(context.Background())
	a.wgp = &sync.WaitGroup{}

	a.assLog = false
	var err error
	// prepare
	for _, v := range m {
		err = v.Prepare(a.Flag)
		a.assert(err, fmt.Sprintf("mod [%v] prepare", v.ID()))
	}

	// init
	err = a.Init()
	a.assert(err, fmt.Sprintf("init main"))

	for _, v := range m {
		err = v.Init(a.Conf, a.rpcSrv)
		a.assert(err, fmt.Sprintf("mod [%v] init", v.ID()))
	}

	// start
	err = a.Start(a.ctx)
	a.assert(err, fmt.Sprintf("main start"))

	for _, v := range m {
		err = v.Start(a.ctx, a.wgp)
		a.assert(err, fmt.Sprintf("mod [%v] start", v.ID()))
	}

	a.writePidFile()
	logger.DisableStdErr()

	if a.Main != nil {
		a.Main()
	}

	a.wgp.Wait()

	a.deletePidFile()

	logger.WARN("Say byebye to the world")

	//time.Sleep(time.Second)
	logger.Sync()
}

// Init
func (a *appCtrl) Init() error {
	// 设置日志标记
	log.SetOutput(os.Stderr)
	log.SetFlags(log.Lshortfile | log.LstdFlags | log.Lmicroseconds)

	// 随机种子
	rand.Seed(time.Now().UnixNano())

	// 命令行解析
	a.assert(a.initFlag(), "init flag")

	logger.INFO(">> APP START")

	// 初始配置
	a.assert(a.initConfig(), "init config")

	// 初始日志
	a.assert(a.initLogger(), "init logger")

	a.assLog = true

	// nmc
	a.assert(a.initNmc(), "init nmc")

	// 初始性能调优
	a.assert(a.initProfile(), "init profile")

	// pid
	a.assert(a.initPidFile(), "init pidfile")

	// 输出基本配置项
	a.showConfig()

	// 初始化网络
	a.assert(a.initRpcSrv(), "init rpc")

	return nil
}

// start
func (a *appCtrl) Start(c context.Context) error {
	a.wgp.Add(1)
	var err error

	if a.rpcSrv != nil {
		err = a.rpcSrv.Serve()
		if err != nil {
			return err
		}
	}

	go func(cx context.Context) {
		tk30Sec := ftime.NewTicker(time.Second * 30)
		tk5Min := ftime.NewTicker(time.Minute * 5)

		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGHUP, syscall.SIGTERM, syscall.SIGINT)

		for {
			select {
			case sig := <-sigCh:
				{
					switch sig {

					case syscall.SIGHUP: // reload config  1
						logger.INFO("SIGHUP reload config")
						param := &pb.SSAppCmd{
							Cmd:    "Reload",
							Source: "signal",
						}
						a.Reload(param)

					case syscall.SIGINT: // app close   2
						logger.INFO("SIGINT stop app")
						a.cancel()

					case syscall.SIGTERM: // app close   15
						logger.INFO("SIGTERM stop app")
						a.cancel()

					default:
						logger.INFO("SIG inv signal:%v", sig)
					}
				}
			case <-cx.Done():
				{
					a.Stop()
					return
				}
			case <-tk30Sec.C:
				{
					//TODO status report
				}
			case <-tk5Min.C:
				{
					runtime.GC()
					logger.INFO("NUM GOROUTINE:%v", runtime.NumGoroutine())
					logger.INFO("GC Summory \n%v", debug.GCSummary())
				}
			}
		}
	}(c)

	go checkStop(a.cancel, os.Getpid())

	return nil
}

// stop
func (a *appCtrl) Stop() error {
	logger.INFO("App will stop")

	for _, v := range a.Modules {
		func(m ModuleItf) {
			err := m.Stop()
			logger.INFO("mod [%v] stop result:%v", m.ID(), err)
		}(v)
	}

	a.wgp.Done()
	logger.INFO("App stop done")
	return nil
}

func (a *appCtrl) Reload(param *pb.SSAppCmd) error {
	logger.INFO("App reload:%v", param.Source)
	c, err := a.loadFunc(nil)
	if err != nil {
		logger.ERR("reload config failed:%v", err)
		return err
	}

	// log lv
	logLv := c.GetInt("loglv")
	logger.INFO("reload set log lv:%v", logLv)
	logger.SetLogLevel(logLv)

	ctx := context.WithValue(context.Background(), "config", c)
	for _, v := range a.Modules {
		go func() {
			err := v.HandleAppCmd(ctx, param)
			logger.INFO("mod [%v] reload result:%v", v.ID(), err)
		}()
	}

	return nil
}

// 命令行解析
func (a *appCtrl) initFlag() error {
	goarg.MustParse(a.Flag)
	if a.Flag.GetBase().Version == true {
		a.showVersion()
		os.Exit(0)
	}

	return nil
}

// 读取配置
func (a *appCtrl) initConfig() error {
	fb := a.Flag.GetBase()
	if fb.Provider == "" {
		fb.Provider = confer.DEF_HCONF_PROVIDER_LOCAL
	} else {
		fb.Provider = confer.DEF_HCONF_PROVIDER_CONFD
	}

	// 配置顺序  -f > -s > a.ConfName > a.Name
	confName := a.ConfName
	confTyp := "toml"
	var serverType string
	if ns := strings.Split(confName, "."); len(ns) > 1 {
		serverType = ns[0]
		confTyp = ns[1]
	}
	switch fb.Provider {
	case confer.DEF_HCONF_PROVIDER_LOCAL:
		if a.ConfName != "" {
			confName = a.ConfName
		} else {
			confName = fmt.Sprintf("%v.json", values.GetAppShard())
		}
	case confer.DEF_HCONF_PROVIDER_CONFD:
		confName = values.GetAppShard()
		if fb.Shard != "" {
			confName = fb.Shard
		}
	}

	var err error

	a.loadFunc = func(res interface{}) (confer.Confer, error) {
		provider := fb.Provider
		param := confer.ConferParam{
			Name: confName,
			Addr: values.GetConfdAddrs(),
			Path: values.GetAppConfDir(),
			Typ:  confTyp,
		}
		c, e := confer.ReadConf(provider, param, res)
		return c, e
	}
	a.Conf, err = a.loadFunc(a.BaseConf)

	if err != nil {
		return err
	}

	cb := a.BaseConf
	var pid *values.Pid
	if cb.Pid != "" {
		pid = values.NewPid("uw", serverType, cb.Pid)
	} else {
		idx, _ := uuid.NewUUID()
		pid = values.NewPid("uw", serverType, idx.String())
	}
	values.SetPid(pid)

	return nil
}

func (a *appCtrl) initLogger() error {
	fb := a.Flag.GetBase()
	cb := a.BaseConf

	logLv := cb.LogLv
	if logLv == 0 {
		logLv = logger.LOGLV_TRACE
	}

	logMode := cb.LogMode
	if logMode == 0 {
		logMode = logger.LOG_OUTPUT_FILE
	}
	if fb.Console {
		logMode = logMode | logger.LOG_OUTPUT_STD
	}

	logPath := values.GetAppLogDir()
	logOssPath := values.GetAppOssLogDir()
	if cb.LogPath != "" {
		logPath = cb.LogPath
	}
	if cb.LogOssPath != "" {
		logOssPath = cb.LogOssPath
	}

	logInterval := cb.LogInterval
	if logInterval == 0 {
		logInterval = logger.LOG_INTERVAL_DAY
	}
	logMaxSize := cb.LogMaxSize
	if logMaxSize == 0 {
		logMaxSize = 500 // log default max size 500MB
	}
	c := &logger.LoggerCfg{
		Name:           values.GetPid().SrvName(),
		Level:          logLv,
		Mode:           logMode,
		Path:           logPath,
		OssPath:        logOssPath,
		Interval:       logInterval,
		Maxsize:        logMaxSize,
		LocalLv:        logLv,
		UdpAddr:        cb.LogUdpAddr,
		LogWechatError: !fb.ErrorWechatNotNotify,
	}

	return logger.InitLog(c)
}

func (a *appCtrl) initNmc() error {
	err := devops.InitNmcAgent()
	if err != nil {
		return err
	}
	return nil
}

func (a *appCtrl) initProfile() error {
	cb := a.BaseConf
	cpuNum := cb.CpuNum
	if cpuNum == 0 || cpuNum >= runtime.NumCPU() {
		cpuNum = runtime.NumCPU()
	}
	nRet := runtime.GOMAXPROCS(cpuNum)
	logger.INFO("app boot init cpu num: %v,%v", cpuNum, nRet)

	if cb.ProfAddr != "" && cb.ProfAddr != "disable" {
		profAddr := cb.ProfAddr
		if cb.ProfAddr == "auto" {
			err, ip, port := devops.GetAddr(devops.DefPortRangeProf, cb.ProfAddr, "prof1")
			if err != nil {
				return err
			}
			profAddr = net.JoinHostPort(ip, strconv.Itoa(port))
		}
		debug.StartPprofSrv(profAddr)
	}

	if cb.StatAddr != "disable" {
		statAddr := cb.StatAddr
		if cb.StatAddr == "" || cb.StatAddr == "auto" {
			err, ip, port := devops.GetAddr(devops.DefPortRangeProf, cb.ProfAddr, "stat1")
			if err != nil {
				return err
			}
			statAddr = net.JoinHostPort(ip, strconv.Itoa(port))
		}
		err := devops.InitStat()
		if err != nil {
			return err
		}
		devops.StartStatSrv(statAddr)
	}

	return nil
}

func (a *appCtrl) initRpcSrv() error {
	rpcAddr := a.BaseConf.RpcAddr
	rpcNet := a.BaseConf.RpcNet
	mode := a.BaseConf.Mode
	if rpcAddr == "disable" {
		return nil
	}
	if mode == "" {
		mode = "rpc"
	}

	if strings.HasPrefix(rpcAddr, "auto") || rpcAddr == "" {
		err, ip, port := devops.GetAddr(devops.DefPortRangeRpc, rpcAddr, "rpc1")
		if err != nil {
			return err
		}
		rpcAddr = net.JoinHostPort(ip, strconv.Itoa(port))
	} else {
		if strings.HasPrefix(rpcAddr, ":") {
			localIp := xnet.GetLocalIP()
			rpcAddr = localIp + rpcAddr
		}
	}

	if rpcNet == "" {
		rpcNet = "tcp"
	}

	if mode == "rpc" {
		a.rpcSrv = srpc.NewServerEasy(rpcNet, rpcAddr)
		err := a.rpcSrv.EnableEtcdRegistry()
		if err != nil {
			return err
		}
		err = a.rpcSrv.EnableInprocRegistry()
		if err != nil {
			return err
		}
	} else if mode == "local" {
		a.rpcSrv = srpc.NewServerEasy(rpcNet, rpcAddr)
		err := a.rpcSrv.EnableInprocRegistry()
		if err != nil {
			return err
		}
	} else {
		a.rpcSrv = srpc.NewServerSingle(rpcNet, rpcAddr)
	}
	s := &modAppCmd{
		BaseRpcService: srpc.BaseRpcService{},
		ctrl:           a,
	}
	sname := "app." + values.GetPid().SrvName() + "." + values.GetPid().SrvIdx()
	err := a.rpcSrv.RegisterName(sname, s, "")
	if err != nil {
		return err
	}

	logger.INFO("RPC PARAM url:%v@%v, base:%v, mode:%v, service:%v",
		rpcNet, rpcAddr, values.GetRpcdBaseId(), mode, sname)
	return nil
}

func (a *appCtrl) showConfig() {
	logger.INFO("=============== show app config ======================")
	logger.INFO(">> AppName:%v", values.GetPid().SrvName())
	logger.INFO(">> AppExecDir:%v", values.GetAppExecDir())
	logger.INFO(">> AppWorkDir:%v", values.GetAppWorkDir())
	logger.INFO(">> AppBinDir:%v", values.GetAppBinDir())
	logger.INFO(">> AppConfDir:%v", values.GetAppConfDir())
	logger.INFO(">> AppLogDir:%v", values.GetAppLogDir())

	logger.INFO(">> CPU:%v  Pid:%v PID:%v", runtime.NumCPU(), os.Getpid(), values.GetPid().PID())

	logger.INFO("======================================================")
}

func (a *appCtrl) showVersion() {
	log.Printf("app version:%v\n", DINF_VER)
	log.Printf("app source:%v\n", DINF_SRC)
	log.Printf("app build time:%v\n", DINF_BTM)
	log.Printf("app build host:%v\n", DINF_HOST)
	log.Printf("app builder info:%v\n", DINF_GO)
}

func (a *appCtrl) initPidFile() error {
	a.pidFile = filepath.Join(values.GetAppPidDir(), fmt.Sprintf("run.%v.pid", values.GetPid().SrvName()))

	err, pid := sysutil.CheckPidFile(a.pidFile)
	if err != nil {
		return err
	}
	if pid != 0 {
		exist, err := sysutil.IsProcessExist(pid)
		if err != nil {
			return err
		}
		if exist == true {
			return fmt.Errorf("process %v exist, check it", pid)
		} else {
			logger.INFO("pid exist but process %v not exist, delete pidfile", pid)
			return sysutil.DeletePidFile(a.pidFile)
		}
	}

	return nil
}

func (a *appCtrl) writePidFile() error {
	var err error
	defer func() {
		if err != nil {
			logger.INFO("write pid failed:%v", err)
		}
	}()

	err, pid := sysutil.CheckPidFile(a.pidFile)
	if err != nil {
		return err
	}
	if pid != 0 {
		exist, err := sysutil.IsProcessExist(pid)
		if err != nil {
			return err
		}
		if exist == true {
			return fmt.Errorf("process %v exist, check it", pid)
		} else {
			err = sysutil.DeletePidFile(a.pidFile)
			if err != nil {
				return err
			}
		}
	}

	err = sysutil.WritePidFile(a.pidFile)
	if err != nil {
		return err
	}

	return nil
}

func (a *appCtrl) deletePidFile() {
	sysutil.DeletePidFile(a.pidFile)
}

func (a *appCtrl) assert(err error, s string) {
	if err != nil {
		if a.ServerInitErr != nil {
			a.ServerInitErr(err)
		}
		logger.FATAL("app boot failed, %v: %v", s, err)
	} else {
		if a.assLog == true {
			logger.INFO("app boot %v done", s)
		}
	}
}

// modAppCmd
func (b *modAppCmd) verify(cname string, r interface{}, args, reply *pb.SSAppCmd) error {
	reply.Cmd = args.Cmd
	if r == nil {
		err := fmt.Errorf("invalid app cmd request:%v", cname)
		logger.ERR(err.Error())
		reply.Emsg = err.Error()
		return err
	}

	return nil
}

func (b *modAppCmd) Stop(ctx context.Context, args, reply *pb.SSAppCmd) error {
	cmd := "Stop"
	request := args.Stop
	if b.verify(cmd, request, args, reply) != nil {
		return nil
	}

	secs := int(request.After)
	if secs < 1 {
		secs = 1
	}

	go func() {
		logger.INFO("APPCMD[%v] app will stop after %v seconds", "Stop")
		time.Sleep(time.Second * time.Duration(secs))
		b.ctrl.cancel()
	}()

	return nil
}

func (b *modAppCmd) Reload(ctx context.Context, args, reply *pb.SSAppCmd) error {
	cmd := "Reload"
	request := args.Reload
	if b.verify(cmd, request, args, reply) != nil {
		return nil
	}

	b.ctrl.Reload(args)

	return nil
}

func (b *modAppCmd) Offline(ctx context.Context, args *pb.SSNull, reply *pb.SSNull) error {
	return nil
}

func (b *modAppCmd) GC(ctx context.Context, args *pb.SSNull, reply *pb.SSNull) error {
	return nil
}

func (b *modAppCmd) Version(ctx context.Context, args *pb.SSNull, reply *pb.SSNull) error {
	return nil
}
