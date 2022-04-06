package registry

import (
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/library/logger"
	"path"
	"reflect"
	"strings"
)

/*
所有接口都必须已经登录
使用updater时必须使用playerHandle.Data()来获取updater
*/

//NewService name: /x/y
//文件加载init()中调用

func NewService(name string, opts *Options) *Service {
	r := &Service{
		Options: opts,
		nodes:   make(map[string]*Node),
		method:  make(map[string]reflect.Value),
	}
	r.name = r.Clean(name)
	return r
}

type Service struct {
	*Options
	name   string
	nodes  map[string]*Node
	method map[string]reflect.Value
}

func (this *Service) Name() string {
	return this.name
}

//Register
func (this *Service) Register(i interface{}, name ...string) error {
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

func (this *Service) RegisterFun(i interface{}, name ...string) error {
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
	fname = this.Clean(fname)
	if fname == "" {
		return errors.New("RegisterFun name empty")
	}
	var proto reflect.Value
	if this.Options.Filter != nil && !this.Options.Filter(proto, v) {
		return fmt.Errorf("RegisterFun filter return false:%v", fname)
	}

	if strings.LastIndex(fname, "/") > 0 {
		return fmt.Errorf("RegisterFun name error:%v", fname)
	}
	if _, ok := this.method[fname]; ok {
		return fmt.Errorf("RegisterFun exist:%v", fname)
	}
	this.method[fname] = v
	this.Options.addRoutePath(this, fname)
	return nil
}

//Register 注册一组handle
func (this *Service) RegisterStruct(i interface{}, name ...string) error {
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
	sname = this.Clean(sname)
	if sname == "" {
		return errors.New("RegisterStruct name empty")
	}
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
		if this.Options.Filter != nil && !this.Options.Filter(v, method.Func) {
			continue
		}
		fname = this.Clean(fname)
		node.method[fname] = method.Func
		this.Options.addRoutePath(this, sname, fname)
	}
	return nil
}

//Match 匹配一个路径
// path : $prefix/$methodName
// path : $prefix/$nodeName/$methodName
func (this *Service) Match(path string) (proto, fn reflect.Value, ok bool) {
	index := len(this.name)
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

func (this *Service) Paths() (r []string) {
	for k, _ := range this.method {
		r = append(r, k)
	}
	for k, node := range this.nodes {
		for n, _ := range node.method {
			r = append(r, path.Join(k, n))
		}
	}

	return
}
