package cosweb

import (
	"fmt"
	"github.com/hwcer/cosgo/library/registry"
	"github.com/hwcer/cosgo/logger"
	"reflect"
	"strings"
)

const (
//RegistryPathName = "_RegistryPathName"
)

/*
通过registry集中注册对象
*/
type RegistryCaller interface {
	Caller(c *Context, fn reflect.Value) interface{}
}

type RegistrySerialize func(ctx *Context, reply interface{})

//NewRegistry 创建新的路由组
// prefix路由前缀
func NewRegistry(prefix string) *Registry {
	r := &Registry{}
	r.prefix = "/" + strings.Trim(prefix, "/")
	r.Registry = registry.New(r.filter)
	r.middleware = make(map[string][]MiddlewareFunc)

	return r
}

type Registry struct {
	*registry.Registry
	prefix     string
	middleware map[string][]MiddlewareFunc
	Caller     func(c *Context, pr reflect.Value, fn reflect.Value) (interface{}, error) //自定义全局消息调用
	Serialize  RegistrySerialize                                                         //消息序列化封装
}

func (r *Registry) filter(pr, fn reflect.Value) bool {
	if !pr.IsValid() {
		_, ok := fn.Interface().(func(*Context) interface{})
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
func (r *Registry) handle(c *Context, next Next) (err error) {
	path := c.Request.URL.Path
	if path == "" || strings.Contains(path, ".") || !strings.HasPrefix(path, r.prefix) {
		return next()
	}
	if r.prefix != "/" {
		path = strings.TrimPrefix(path, r.prefix)
	}

	route, pr, fn, ok := r.Registry.Match(path)
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
		r.Serialize(c, reply)
	} else {
		c.JSON(reply)
	}
	return
}

func (r *Registry) caller(c *Context, pr, fn reflect.Value) (reply interface{}, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
			logger.Error("%v", err)
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
	route := r.Registry.Route(name)
	if len(middleware) > 0 {
		s := route.Name()
		r.middleware[s] = append(r.middleware[s], middleware...)
	}
	return route
}

//Handle 注册服务器
func (r *Registry) Handle(s *Server, method ...string) {
	for _, k := range r.Registry.Nodes() {
		var arr []string
		if r.prefix != "/" {
			arr = append(arr, r.prefix)
		}
		if k != "/" {
			arr = append(arr, k)
		}
		arr = append(arr, "/*")

		route := strings.Join(arr, "")
		s.Register(route, r.handle, method...)
	}
}
