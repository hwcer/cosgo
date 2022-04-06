package cosrpc

import (
	"context"
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/library/registry"
	"github.com/hwcer/cosgo/values"
	"github.com/smallnest/rpcx/server"
	"reflect"
	"strings"
)

// 通过registry集中注册对象

type RegistryCaller interface {
	Caller(fn reflect.Value) interface{}
}
type RegistryFilter func(pr, fn reflect.Value) bool
type RegistrySerialize func(reply interface{}) error

type RegistryArgs struct {
	methodName  string
	serviceName string
}

//NewRegistry 创建新的路由组
// prefix路由前缀
func NewRegistry(prefix string) *Registry {
	r := &Registry{}
	r.Registry = registry.New(prefix, r.filter)
	return r
}

type Registry struct {
	*registry.Registry
	Caller    func(ctx context.Context, pr reflect.Value, fn reflect.Value) (interface{}, error) //自定义全局消息调用
	Filter    RegistryFilter
	Serialize RegistrySerialize //消息序列化封装
}

func (r *Registry) filter(pr, fn reflect.Value) bool {
	if r.Filter != nil {
		return r.Filter(pr, fn)
	}
	if !pr.IsValid() {
		_, ok := fn.Interface().(func(*RegistryArgs) interface{})
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

//handle rpcx 入口
func (r *Registry) handle(c *server.Context) (err error) {
	route, ok := r.Get(c.ServicePath())
	if !ok {
		return errors.New("service not exist")
	}

	path := c.Request.URL.Path
	if path == "" || strings.Contains(path, ".") {
		return next()
	}
	route, pr, fn, ok := r.Registry.Match(path[r.Index():])
	if !ok {
		return next()
	}
	name := route.Name()
	if err, ok = c.doMiddleware(r.middleware[name]); err != nil || !ok {
		return
	}

	var reply interface{}
	reply, err = r.caller(c, pr, fn)

	if err != nil {
		return
	}
	if r.Serialize != nil {
		return r.Serialize(c, reply)
	} else {
		return c.JSON(reply)
	}
}

func (r *Registry) caller(c *Context, pr, fn reflect.Value) (reply interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
			//logger.Error("%v", err)
		}
	}()
	if r.Caller != nil {
		return r.Caller(c, pr, fn)
	}
	if !pr.IsValid() {
		f, _ := fn.Interface().(func(c *Context) interface{})
		reply = f(c)
	} else if s, ok := pr.Interface().(RegistryCaller); ok {
		reply = s.Caller(c, fn)
	} else {
		ret := fn.Call([]reflect.Value{pr, reflect.ValueOf(c)})
		reply = ret[0].Interface()
	}
	return
}

func (r *Registry) Route(name string, middleware ...MiddlewareFunc) *registry.Route {
	route := r.Registry.Namespace(name)
	if len(middleware) > 0 {
		s := route.Name()
		r.middleware[s] = append(r.middleware[s], middleware...)
	}
	return route
}

//Handle 注册服务器
func (r *Registry) Handle(s *Server, method ...string) {
	for _, path := range r.Paths() {
		s.Register(path+"/*", r.handle, method...)
	}
}
