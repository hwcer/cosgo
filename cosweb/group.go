package cosweb

import (
	"cosgo/logger"
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
	nodes  map[string]*GroupNode
	caller GroupCaller
}

//GroupNode 节点，每个节点对应一个容器以及容器下面的所有接口
type GroupNode struct {
	proto      reflect.Value
	value      map[string]reflect.Value
	middleware []MiddlewareFunc
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

//Route 将Route添加到服务
func (g *Group) Route(s *Server, prefix string, method ...string) *Group {
	arr := []string{strings.TrimSuffix(prefix, "/"), ":" + iGroupRoutePath, ":" + iGroupRouteName}
	r := strings.Join(arr, "/")
	s.Register(r, g.handle, method...)
	return g
}

//Caller 设置Group的caller
func (g *Group) Caller(f GroupCaller) *Group {
	g.caller = f
	return g
}

//Register 注册一组handle，重名忽略
func (g *Group) Register(handle interface{}, middleware ...MiddlewareFunc) *Group {
	handleType := reflect.TypeOf(handle)
	if handleType.Kind() != reflect.Ptr {
		panic("Group Register error:handle not pointer")
	}
	name := strFirstToLower(handleType.Elem().Name())
	if _, ok := g.nodes[name]; ok {
		panic(fmt.Sprintf("Group Register error:%v exist", name))
	}
	node := NewGroupNode(handle)
	node.middleware = middleware
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
	return g
}

//handle 路由入口
func (g *Group) handle(c *Context) (err error) {
	path := c.Param(iGroupRoutePath)
	name := c.Param(iGroupRouteName)
	node := g.nodes[path]
	if node == nil {
		return nil
	}
	//node.middleware
	if len(node.middleware) > 0 {
		c.middleware = append(c.middleware, node.middleware...)
		c.next()
		if c.Aborted() {
			return nil
		}
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
