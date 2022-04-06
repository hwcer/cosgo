package cosrpc

import (
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/library/registry"
	"github.com/hwcer/cosgo/utils"
	"github.com/smallnest/rpcx/server"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// 通过registry集中注册对象

type XServerRegistryCaller interface {
	Caller(c *server.Context, fn reflect.Value) interface{}
}

type XServerRegistrySerialize func(c *server.Context, reply interface{}) error

type RpcRegister interface {
	Stop() error
	Start() error
	Register(name string, i interface{}, metadata string) (err error)
}

func NewXServer(opts *registry.Options) *XServer {
	r := &XServer{}
	if opts == nil {
		opts = registry.NewOptions()
	}
	if opts.Filter == nil {
		opts.Filter = r.filter
	}
	r.Registry = registry.New(opts)
	return r
}

type XServer struct {
	*registry.Registry
	Caller      func(c *server.Context, pr reflect.Value, fn reflect.Value) (interface{}, error) //自定义全局消息调用
	Serialize   XServerRegistrySerialize                                                         //消息序列化封装
	rpcServer   *server.Server
	rpcRegister RpcRegister
}

func (this *XServer) filter(pr, fn reflect.Value) bool {
	if !pr.IsValid() {
		_, ok := fn.Interface().(func(*server.Context) interface{})
		return ok
	}
	t := fn.Type()
	if t.NumIn() != 2 {
		return false
	}
	if t.NumOut() != 1 {
		return false
	}
	return true
}

//handle cosweb入口
func (this *XServer) handle(c *server.Context) (err error) {
	urlPath := this.Clean(c.ServicePath(), c.ServiceMethod())
	route, ok := this.Match(urlPath)
	if !ok {
		return errors.New("ServicePath not exist")
	}
	pr, fn, ok := route.Match(urlPath)
	if !ok {
		return errors.New("ServiceMethod not exist")
	}

	var reply interface{}
	reply, err = this.caller(c, pr, fn)

	if err != nil {
		return
	}
	if this.Serialize != nil {
		return this.Serialize(c, reply)
	} else {
		return c.Write(reply)
	}
}

func (this *XServer) caller(c *server.Context, pr, fn reflect.Value) (reply interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
			//logger.Error("%v", err)
		}
	}()
	if this.Caller != nil {
		return this.Caller(c, pr, fn)
	}
	if !pr.IsValid() {
		f, _ := fn.Interface().(func(c *server.Context) interface{})
		reply = f(c)
	} else if s, ok := pr.Interface().(XServerRegistryCaller); ok {
		reply = s.Caller(c, fn)
	} else {
		ret := fn.Call([]reflect.Value{pr, reflect.ValueOf(c)})
		reply = ret[0].Interface()
	}
	return
}

func (this *XServer) Server() *server.Server {
	return this.rpcServer
}

func (this *XServer) Route(name string) *registry.Service {
	route := this.Registry.Service(name)
	return route
}

func (this *XServer) Services() (s []string) {
	this.Registry.Range(func(name string, _ *registry.Service) bool {
		servicePath := strings.TrimPrefix(name, "/")
		s = append(s, servicePath)
		return true
	})
	return
}

func (this *XServer) Start(address *url.URL, register RpcRegister) (err error) {
	if err = register.Start(); err != nil {
		return
	}
	this.rpcServer = server.NewServer()
	this.rpcServer.DisableHTTPGateway = true
	this.Registry.Range(func(name string, route *registry.Service) bool {
		servicePath := strings.Trim(name, "/")
		register.Register(servicePath, nil, "")
		for _, serviceMethod := range route.Paths() {
			this.rpcServer.AddHandler(servicePath, serviceMethod, this.handle)
		}
		return true
	})

	scheme := address.Scheme
	if scheme == "" {
		scheme = "tcp"
	}
	err = utils.Timeout(time.Second, func() error {
		return this.rpcServer.Serve(scheme, address.Host)
	})
	if err == utils.ErrorTimeout {
		err = nil
	}
	return
}

func (this *XServer) Close() error {
	this.rpcServer.Shutdown(nil)
	this.rpcRegister.Stop()
	return nil
}
