package app

import (
	"context"
	"icefire/confer"
	"icefire/pb"
	"icefire/srpc"
	"sync"
)

// flag接口
type FlagItf interface {
	GetBase() *AppBaseFlag
}

// 模块接口
type ModuleItf interface {
	ID() string
	Prepare(f FlagItf) error
	Init(c confer.Confer, s *srpc.SRpcServer) error
	Start(cx context.Context, wgp *sync.WaitGroup) error
	Stop() error
	HandleAppCmd(cx context.Context, param *pb.SSAppCmd) error
}

func SetMain(m func()) {
	defaultCtrl.Main = m
}

func SetServerInitError(m func(err error)) {
	defaultCtrl.ServerInitErr = m
}

/**
 * 应用程序启动
 * @param name 应用程序名
 * @param defConf 默认配置文件名
 * @param iFlag  自定义的flag配置，可以为空
 * @param m 需注册的模块
 */
func Run(defConf string, iFlag FlagItf, m ...ModuleItf) {
	if iFlag == nil {
		iFlag = new(baseFlagImpl)
	}
	defaultCtrl.Run(defConf, iFlag, m...)
}

func Exit(source string) {
	defaultCtrl.cancel()
}
