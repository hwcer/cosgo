package cosrpc

import (
	"github.com/hwcer/cosgo/utils"
	"github.com/smallnest/rpcx/server"
	"net/url"
	"strings"
	"time"
)

func NewXServer() *XServer {
	return &XServer{}
}

type XServer struct {
	rpcxServer   *server.Server
	rpcxServices []*service
	rpcxRegister Register
}

type service struct {
	rcvr        interface{}
	metadata    string
	methodName  string
	servicePath string
}

type Register interface {
	Stop() error
	Start() error
}

func (this *XServer) Register(name string, rcvr interface{}, metadata ...string) {
	this.RegisterFunc(name, "", rcvr, metadata...)
}

func (this *XServer) RegisterFunc(servicePath, methodName string, rcvr interface{}, metadata ...string) {
	s := &service{
		rcvr:        rcvr,
		methodName:  methodName,
		servicePath: servicePath,
	}
	if len(metadata) > 0 {
		s.metadata = strings.Join(metadata, "&")
	}
	this.rpcxServices = append(this.rpcxServices, s)
}

//address 使用内网或者外网地址，不能使用127.0.0.1,localhost之类的外网无法访问
func (this *XServer) Start(address *url.URL, register Register) (err error) {
	//server
	if err = register.Start(); err != nil {
		return
	}
	this.rpcxServer = server.NewServer()
	this.rpcxServer.DisableHTTPGateway = true
	this.rpcxServer.Plugins.Add(register)
	this.rpcxRegister = register
	for _, v := range this.rpcxServices {
		metadata := v.metadata
		if v.methodName != "" {
			err = this.rpcxServer.RegisterFunctionName(v.servicePath, v.methodName, v.rcvr, metadata)
		} else {
			err = this.rpcxServer.RegisterName(v.servicePath, v.rcvr, metadata)
		}
		if err != nil {
			return
		}
	}
	var scheme = address.Scheme
	if scheme == "" {
		scheme = "tcp"
	}
	err = utils.Timeout(time.Second, func() error {
		return this.rpcxServer.Serve(scheme, address.Host)
	})
	if err == utils.ErrorTimeout {
		err = nil
	}
	return
}

func (this *XServer) Close() (errs []error) {

	var err error
	if err = this.rpcxServer.Shutdown(nil); err != nil {
		errs = append(errs, err)
	}
	if err = this.rpcxRegister.Stop(); err != nil {
		errs = append(errs, err)
	}
	return
}

func (this *XServer) Size() int {
	return len(this.rpcxServices)
}
