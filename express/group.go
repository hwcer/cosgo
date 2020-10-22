package express

import (
	"cosgo/logger"
	"reflect"
)

const (
	iGroupRoutePath = "_GROUP_PATH"
	iGroupRouteName = "_GROUP_NAME"
)

var typeOfContext = reflect.TypeOf(&Context{})

//Group  namsespace代表了一个虚拟目录，目录下有很多method
//可以使用 /path/$Group/$value 的格式访问
type Group struct {
	nodes map[string]*GroupNode
}

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

//Handle 路由入口
func (this *Group) handler(c *Context) error {
	path := c.Param(iGroupRoutePath)
	name := c.Param(iGroupRouteName)
	nsp := this.nodes[path]
	if nsp == nil {
		return nil
	}
	//原始方法
	if f, ok := nsp.method[name]; ok {
		return f(c)
	}
	//反射方法
	name = strFirstToUpper(name)
	var ok bool
	var method reflect.Value
	if method, ok = nsp.value[name]; !ok {
		return nil
	}
	ret := method.Call([]reflect.Value{nsp.proto, reflect.ValueOf(c)})
	if ret[0].IsNil() {
		return nil
	} else {
		return ret[0].Interface().(error)
	}
}

//Register 注册一组handle，重名忽略
func (this *Group) Register(handle interface{}) {
	handleType := reflect.TypeOf(handle)
	if handleType.Kind() != reflect.Ptr {
		logger.Error("Group Register error:handle not pointer")
		return
	}
	name := handleType.Elem().Name()
	if _, ok := this.nodes[name]; ok {
		logger.Error("Group Register error:%v exist", name)
	}
	node := NewGroupNode(handle)

	this.nodes[name] = node
	//fmt.Printf("Register:%v\n",Group.name)
	for m := 0; m < handleType.NumMethod(); m++ {
		method := handleType.Method(m)
		methodType := method.Type
		methodName := method.Name
		//fmt.Println("打印Method", methodName, methodType)
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

		node.value[methodName] = method.Func

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
