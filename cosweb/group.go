package cosweb

import (
	"cosgo/logger"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

const (
	iGroupRoutePath = "_GroupRoutePath"
	iGroupRouteName = "_GroupRouteName"
)

var typeOfContext = reflect.TypeOf(&Context{})

//Group 使用反射集中注册方法
//可以使用 /prefix/$Group/$value 的格式访问
type Group struct {
	nodes      map[string]*GroupNode
	caller     GroupCaller
	middleware []MiddlewareFunc
}

//GroupNode 节点，每个节点对应一个容器以及容器下面的所有接口
type GroupNode struct {
	proto reflect.Value
	value map[string]reflect.Value
}

//GroupCaller 建议为每个注册的struct对象封装一个caller方法可以避免使用 reflect.Value.Call()方法
type GroupCaller func(reflect.Value, reflect.Value, *Context) error

//NewGroup 创建新的路由组
func NewGroup() *Group {
	return &Group{
		nodes: make(map[string]*GroupNode),
	}
}

//NewGroupNode NewGroupNode
func NewGroupNode(handle interface{}) *GroupNode {
	return &GroupNode{
		proto: reflect.ValueOf(handle),
		value: make(map[string]reflect.Value),
	}
}

//Use 使用中间件，只针对GROUP下的API
func (g *Group) Use(m ...MiddlewareFunc) {
	g.middleware = append(g.middleware, m...)
}

//Route 将Route添加到服务
func (g *Group) Route(s *Server, prefix string, method ...string) {
	arr := []string{strings.TrimSuffix(prefix, "/"), ":" + iGroupRoutePath, ":" + iGroupRouteName}
	r := strings.Join(arr, "/")
	s.Register(r, g.handle, method...)
}

//SetCaller 设置Group的caller
func (g *Group) SetCaller(f GroupCaller) {
	g.caller = f
}

//Register 注册一组handle，重名忽略
func (g *Group) Register(handle interface{}) error {
	handleType := reflect.TypeOf(handle)
	if handleType.Kind() != reflect.Ptr {
		return errors.New("Group Register error:handle not pointer")
	}
	name := strFirstToLower(handleType.Elem().Name())
	if _, ok := g.nodes[name]; ok {
		return errors.New(fmt.Sprintf("Group Register error:%v exist", name))
	}
	node := NewGroupNode(handle)
	g.nodes[name] = node
	//logger.Debug("Register:%v\n", name)
	for m := 0; m < handleType.NumMethod(); m++ {
		method := handleType.Method(m)
		methodType := method.Type
		methodName := method.Name
		//logger.Debug("Method,name:%v,type:%v", methodName, methodType)
		// value must be exported.
		if method.PkgPath != "" {
			logger.Debug("Register value PkgPath Not End,value:%v.%v(),PkgPath:%v", name, methodName, method.PkgPath)
			continue
		}
		if !isExported(methodName) {
			logger.Debug("Register value Can't Exported,value:%v.%v()", name, methodName)
			continue
		}
		// value needs four ins: receiver, context.Context, *args, *reply.
		if methodType.NumIn() != 2 || methodType.NumOut() != 1 {
			logger.Debug("Register value args num or return num error,value:%v.%v()", name, methodName)
			continue
		}
		// First arg must be context.Context
		ctxType := methodType.In(1)
		if !ctxType.ConvertibleTo(typeOfContext) {
			logger.Debug("Register value args error,value:%v.%v()", name, methodName)
			continue
		}
		////
		//outType := methodType.Out(0)
		//if !outType.ConvertibleTo(typeOfMessage) {
		//	logger.Debug("Register value return error,value:%v.%v()\n", name, methodName)
		//	continue
		//}

		node.value[strFirstToLower(methodName)] = method.Func

	}
	return nil
}

//handle 路由入口
func (g *Group) handle(c *Context) (err error) {
	//group middleware
	c.doMiddleware(g.middleware)
	if c.Aborted() {
		return nil
	}
	path := c.Get(iGroupRoutePath, RequestDataTypePath)
	name := c.Get(iGroupRouteName, RequestDataTypePath)
	node := g.nodes[path]
	if node == nil {
		return nil
	}

	//反射方法
	var ok bool
	var method reflect.Value
	if method, ok = node.value[name]; !ok {
		return nil
	}
	if g.caller != nil {
		return g.caller(node.proto, method, c)
	} else {
		ret := method.Call([]reflect.Value{node.proto, reflect.ValueOf(c)})
		if !ret[0].IsNil() {
			err = ret[0].Interface().(error)
		}
	}
	return
}
