package cosweb

import (
	"cosgo/logger"
	"reflect"
	"strings"
)

const (
	iGroupRoutePath = "_GroupRoutePath"
	iGroupRouteName = "_GroupRouteName"
)

var typeOfContext = reflect.TypeOf(&Context{})

//Group 使用反射集中注册方法
//可以使用 /path/$Group/$value 的格式访问
type Group struct {
	nodes  map[string]*GroupNode
	caller GroupCaller
	prefix string
}

//proto method ctx
//建议为每个注册的struct对象封装一个caller方法可以避免使用 reflect.Value.Call()方法
type GroupCaller func(reflect.Value, reflect.Value, *Context) error

type GroupNode struct {
	proto  reflect.Value
	value  map[string]reflect.Value
	method GroupHandler
}
type GroupHandler map[string]HandlerFunc

func NewGroup() *Group {
	return &Group{
		nodes: make(map[string]*GroupNode),
	}
}

func NewGroupNode(handle interface{}) *GroupNode {
	return &GroupNode{
		proto:  reflect.ValueOf(handle),
		value:  make(map[string]reflect.Value),
		method: make(GroupHandler),
	}
}

func (this *Group) Route(prefix string) string {
	arr := []string{strings.TrimSuffix(prefix, "/"), ":" + iGroupRoutePath, ":" + iGroupRouteName}
	r := strings.Join(arr, "/")
	return r
}

func (this *Group) Caller(f GroupCaller) *Group {
	this.caller = f
	return this
}

//Handle 路由入口
func (this *Group) handler(c *Context) (err error) {
	path := c.Param(iGroupRoutePath)
	name := c.Param(iGroupRouteName)
	node := this.nodes[path]
	if node == nil {
		return nil
	}
	//原始方法
	if f, ok := node.method[name]; ok {
		return f(c)
	}
	//反射方法
	var ok bool
	var method reflect.Value
	if method, ok = node.value[name]; !ok {
		return nil
	}
	if this.caller != nil {
		return this.caller(node.proto, method, c)
	} else {
		ret := method.Call([]reflect.Value{node.proto, reflect.ValueOf(c)})
		if !ret[0].IsNil() {
			err = ret[0].Interface().(error)
		}
	}
	return
}

//Register 注册一组handle，重名忽略
func (this *Group) Register(handle interface{}) {
	handleType := reflect.TypeOf(handle)
	if handleType.Kind() != reflect.Ptr {
		logger.Error("Group Register error:handle not pointer")
		return
	}
	name := strFirstToLower(handleType.Elem().Name())
	if _, ok := this.nodes[name]; ok {
		logger.Error("Group Register error:%v exist", name)
		return
	}
	node := NewGroupNode(handle)
	this.nodes[name] = node
	logger.Debug("Register:%v\n", name)
	for m := 0; m < handleType.NumMethod(); m++ {
		method := handleType.Method(m)
		methodType := method.Type
		methodName := method.Name
		logger.Debug("Method,name:%v,type:%v", methodName, methodType)
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
	//MAP MOTHOD
	if handleType.Elem().Kind() == reflect.Map {
		h, ok := handle.(GroupHandler)
		if ok {
			for k, f := range h {
				node.method[k] = f
			}
		}
	}
}
