package registry

import (
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/library/logger"
	"reflect"
	"strings"
)

/*
所有接口都必须已经登录
使用updater时必须使用playerHandle.Data()来获取updater
*/

//NewRoute name: /x/y
//文件加载init()中调用

func NewRoute(registry *Registry, name string) *Route {
	r := &Route{
		Prefix:   NewPrefix(name),
		nodes:    make(map[string]*Node),
		method:   make(map[string]reflect.Value),
		registry: registry,
	}
	return r
}

type Route struct {
	*Prefix
	nodes    map[string]*Node
	method   map[string]reflect.Value
	registry *Registry
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
	fname = this.registry.Format(fname)
	var proto reflect.Value
	if this.registry.filter != nil && !this.registry.filter(proto, v) {
		return fmt.Errorf("RegisterFun filter return false:%v", fname)
	}

	if strings.LastIndex(fname, "/") > 0 {
		return fmt.Errorf("RegisterFun name error:%v", fname)
	}
	if _, ok := this.method[fname]; ok {
		return fmt.Errorf("RegisterFun exist:%v", fname)
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
	node := NewNode(sname, v)
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

func (this *Route) Match(path string) (proto, fn reflect.Value, ok bool) {
	index := this.Index()
	if index > 0 && !strings.HasPrefix(path, this.name) {
		return
	}

	name := path[index:]
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
	if fn, ok = node.method[name[lastIndex:]]; !ok {
		return
	}

	proto = node.i
	return
}

func (r *Route) Range(prefix string, fn RegistryRangeHandle) (err error) {
	rp := prefix + r.Name()
	for k, node := range r.nodes {
		if err = node.Range(rp+k, fn); err != nil {
			return
		}
	}
	for k, m := range r.method {
		if err = fn(rp+k, m); err != nil {
			return
		}
	}
	return nil
}
