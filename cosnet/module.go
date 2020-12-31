package cosnet

import (
	"cosgo/app"
	"errors"
	"sync"
)

//使用APP MODULE方式统一管理服务器
//Id 可当配置地址使用
func New(id, addr string, handler MsgHandler) *Module {
	mod := &Module{Addr: addr, Handler: handler}
	mod.DefModule.Id = id
	app.Use(mod)
	return mod
}

type Module struct {
	app.DefModule
	Addr    string //默认地址
	Flag    string //启动参数，允许在启动参数中定义Addr,高优先级
	Server  Server
	Handler MsgHandler
}

//接口函数禁止调用
//自动通过config加载配置
func (this *Module) Init() (err error) {
	if this.Flag != "" {
		addr := app.Config.GetString(this.Flag)
		if addr != "" {
			this.Addr = addr
		}
	}
	if this.Addr == "" {
		return errors.New("Server Addr empty")
	}
	this.Server, err = NewServer(this.Addr, MsgTypeMsg, this.Handler)
	return
}

//接口函数禁止调用
func (this *Module) Start(wgp *sync.WaitGroup) (err error) {
	wgp.Add(1)
	return this.Server.Start()
}

//接口函数禁止调用
func (this *Module) Close(wgp *sync.WaitGroup) (err error) {
	this.Server.Close()
	wgp.Done()
	return
}
