package cosweb

import (
	"cosgo/apps"
	"crypto/tls"
	"errors"
	"sync"
)

//使用APP MODULE统一管理服务器
//Id 可当配置地址使用
func New(id, addr string, handle interface{}, middleware ...MiddlewareFunc) *module {
	mod := &module{Id: id, Addr: addr}

	mod.Prefix = "*"
	mod.Handle = handle
	mod.Middleware = middleware
	apps.Use(mod)
	return mod
}

type module struct {
	Id       string
	TLS      uint8  //启用TLS，，0-关闭，1-配置路径（必须配置tls.key  tls.pem 路径）2-自动证书
	Addr     string //默认地址
	Flag     string //启动参数，允许在启动参数中定义Addr,高优先级
	Server   *Engine
	Register func(*Engine)

	Prefix     string
	Handle     interface{}
	Middleware []MiddlewareFunc

	//Handler IMsgHandler
}

func (this *module) ID() string {
	return this.Id
}

//接口函数禁止调用
//自动通过config加载配置
func (this *module) Init() (err error) {
	if this.Flag != "" {
		addr := apps.Config.GetString(this.Flag)
		if addr != "" {
			this.Addr = addr
		}
	}
	if this.Addr == "" {
		return errors.New("Express Server Addr empty")
	}
	var tlsConfig *tls.Config
	if this.TLS == 1 {
		tlspem := apps.Config.GetString("tls.pem")
		tlskey := apps.Config.GetString("tls.key")
		if tlspem == "" || tlskey == "" {
			return errors.New("Express Server TLS File empty")
		}
		var cert tls.Certificate
		cert, err = tls.LoadX509KeyPair(tlspem, tlskey)
		tlsConfig = &tls.Config{Certificates: []tls.Certificate{cert}}
	}

	this.Server = NewServer(this.Addr, tlsConfig)
	if this.Handle != nil {
		this.Server.Group(this.Prefix, this.Handle, this.Middleware...)
	}
	if this.Register != nil {
		this.Register(this.Server)
	}
	return
}

//接口函数禁止调用
func (this *module) Start(wgp *sync.WaitGroup) (err error) {
	wgp.Add(1)
	return this.Server.Start()
}

//接口函数禁止调用
func (this *module) Close(wgp *sync.WaitGroup) (err error) {
	err = this.Server.Close()
	wgp.Done()
	return
}
