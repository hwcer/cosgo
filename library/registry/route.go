package registry

import (
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/logger"
	"reflect"
	"strings"
)

/*
所有接口都必须已经登录
使用updater时必须使用playerHandle.Data()来获取updater
*/

//NewRoute name: /x/y
//文件加载init()中调用

func newNode(name string, i reflect.Value) *Node {
	return &Node{name: name, i: i, method: make(map[string]reflect.Value)}
}

func NewRoute(registry *Registry, name string) *Route {
	name = "/" + strings.Trim(name, "/")
	r := &Route{
		name:     name,
		nodes:    make(map[string]*Node),
		method:   make(map[string]reflect.Value),
		registry: registry,
	}
	return r
}

type Node struct {
	i      reflect.Value
	name   string
	method map[string]reflect.Value
}

type Route struct {
	name     string
	nodes    map[string]*Node
	method   map[string]reflect.Value
	registry *Registry
}

func (this *Route) Name() string {
	return this.name
}

//Register
func (this *Route) Register(i interface{}, name ...string) error {
	v := reflect.ValueOf(i)
	var kind reflect.Kind
	if v.Kind() == reflect.Ptr {
		kind = v.Elem().Kind()
	} else {
		kind = v.Kind()
	}
	switch kind {
	case reflect.Func:
		return this.RegisterFun(v, name...)
	case reflect.Struct:
		return this.RegisterStruct(v, name...)
	default:
		return fmt.Errorf("未知的注册类型：%v", v.Kind())
	}
}

func (this *Route) RegisterFun(i interface{}, name ...string) error {
	v := ValueOf(i)
	if v.Kind() != reflect.Func {
		return errors.New("RegisterFun fn type must be reflect.Func")
	}
	var fname string
	if len(name) > 0 {
		fname = name[0]
	} else {
		fname = FuncName(v)
	}
	fname = this.registry.Format(strings.Trim(fname, "/"))
	var proto reflect.Value
	//logger.Debug("RegisterFun:%v", fname)
	if this.registry.filter != nil && !this.registry.filter(proto, v) {
		return fmt.Errorf("RegisterFun filter return false:%v", fname)
	}

	if strings.Contains(fname, "/") {
		return fmt.Errorf("RegisterFun name error:%v", name)
	}
	if _, ok := this.method[fname]; ok {
		return fmt.Errorf("RegisterFun exist:%v", name)
	}
	this.method[fname] = v
	return nil
}

//Register 注册一组handle
func (this *Route) RegisterStruct(i interface{}, name ...string) error {
	v := ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return errors.New("RegisterStruct handle type must be reflect.Struct")
	}
	if v.Elem().Kind() != reflect.Struct {
		return errors.New("RegisterStruct handle type must be reflect.Struct")
	}
	handleType := v.Type()
	var sname string
	if len(name) > 0 {
		sname = name[0]
	} else {
		sname = handleType.Elem().Name()
	}
	sname = this.registry.Format(sname)
	if _, ok := this.nodes[sname]; ok {
		return fmt.Errorf("RegisterStruct name exist:%v", sname)
	}
	node := newNode(sname, v)
	this.nodes[sname] = node
	//logger.Debug("Watch:%v\n", sname)
	for m := 0; m < handleType.NumMethod(); m++ {
		method := handleType.Method(m)
		//methodType := method.Type
		fname := method.Name
		//logger.Debug("Watch,sname:%v,type:%v", fname, methodType)
		// value must be exported.
		if method.PkgPath != "" {
			logger.Debug("Watch value PkgPath Not End,value:%v.%v(),PkgPath:%v", sname, fname, method.PkgPath)
			continue
		}
		if !IsExported(fname) {
			logger.Debug("Watch value Can't Exported,value:%v.%v()", sname, fname)
			continue
		}
		// value needs four ins: receiver, context.Ctx, *data, *reply.
		//if methodType.NumIn() != 2 || methodType.NumOut() != 1 {
		//	logger.Debug("Watch value data num or return num error,value:%v.%v()", sname, fname)
		//	continue
		//}
		if this.registry.filter != nil && !this.registry.filter(v, method.Func) {
			continue
		}
		fname = this.registry.Format(fname)
		node.method[fname] = method.Func
	}
	return nil
}

func (this *Route) match(path string) (proto, fn reflect.Value, ok bool) {
	if !strings.HasPrefix(path, this.name) {
		return
	}
	var prefix = this.name
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}

	name := strings.TrimPrefix(path, prefix)
	if fn, ok = this.method[name]; ok {
		return
	}
	lastIndex := strings.LastIndex(name, "/")
	if lastIndex <= 0 {
		return
	}
	var node *Node
	if node, ok = this.nodes[name[0:lastIndex]]; !ok {
		return
	}
	if fn, ok = node.method[name[lastIndex+1:]]; !ok {
		return
	}

	proto = node.i
	return
}
