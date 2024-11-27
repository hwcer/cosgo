package registry

import (
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"
)

/*
所有接口都必须已经登录
*/

//NewService name: /x/y
//文件加载init()中调用

//type FilterEventType int8
//
//const (
//	FilterEventTypeFunc FilterEventType = iota
//	FilterEventTypeMethod
//	FilterEventTypeStruct
//)

type Handler interface {
	Filter(*Node) bool
}

func NewService(name string, router *Router) *Service {
	r := &Service{
		nodes:  make(map[string]*Node),
		router: router,
	}
	r.prefix = Join(name)
	if len(r.prefix) > 1 {
		r.name = r.prefix[1:]
	}
	//r.Formatter = strings.ToLower
	return r
}

type Service struct {
	name   string // a/b
	prefix string //  /a/b
	nodes  map[string]*Node
	//events    map[FilterEventType]func(*Node) bool
	router  *Router
	Handler Handler //自定义 Filter等方法
	//Formatter func(string) string //格式化对象和方法名，默认强制小写
}

//func (this *Service) On(t FilterEventType, l func(*Node) bool) {
//	if this.events == nil {
//		this.events = make(map[FilterEventType]func(*Node) bool)
//	}
//	this.events[t] = l
//}
//
//func (this *Service) Emit(t FilterEventType, node *Node) bool {
//	if this.events == nil {
//		return true
//	}
//	filter := this.events[t]
//	if filter != nil && !filter(node) {
//		return false
//	}
//	return true
//}

func (this *Service) Name() string {
	return this.name
}

func (this *Service) Prefix() string {
	return this.prefix
}
func (this *Service) Merge(s *Service) (err error) {
	if s.Handler != nil {
		this.Handler = s.Handler
	}
	for k, v := range s.nodes {
		node := &Node{name: v.name, value: v.value, binder: v.binder, Service: this}
		this.nodes[k] = node
		if err = this.router.Register(node, node.Route()); err != nil {
			return
		}
	}
	return
}

// Register 服务注册
func (this *Service) Register(i interface{}, prefix ...string) error {
	v := reflect.ValueOf(i)
	var kind reflect.Kind
	if v.Kind() == reflect.Ptr {
		kind = v.Elem().Kind()
	} else {
		kind = v.Kind()
	}
	switch kind {
	case reflect.Func:
		return this.RegisterFun(v, prefix...)
	case reflect.Struct:
		return this.RegisterStruct(v, prefix...)
	default:
		return fmt.Errorf("未知的注册类型：%v", v.Kind())
	}
}

func (this *Service) filter(node *Node) bool {
	if this.Handler == nil {
		return true
	}
	return this.Handler.Filter(node)
}

func (this *Service) format(serviceName, methodName string, prefix ...string) string {
	serviceName = Formatter(serviceName)
	methodName = Formatter(methodName)
	if len(prefix) == 0 {
		return Join(serviceName, methodName)
	}

	p := Join(prefix...)
	var name string
	if serviceName == "" {
		name = methodName
	} else {
		name = path.Join(serviceName, methodName)
	}
	p = strings.Replace(p, "%v", name, -1)
	p = strings.Replace(p, "%s", serviceName, -1)
	p = strings.Replace(p, "%m", methodName, -1)
	return p
}

func (this *Service) RegisterFun(i interface{}, prefix ...string) error {
	v := ValueOf(i)
	if v.Kind() != reflect.Func {
		return errors.New("RegisterFun fn type must be reflect.Func")
	}

	name := this.format("", FuncName(v), prefix...)
	//if name == "" {
	//	return errors.New("RegisterFun name empty")
	//}
	node := &Node{name: name, value: v, Service: this}
	if !this.filter(node) {
		return fmt.Errorf("RegisterFun filter return false:%v", name)
	}

	if _, ok := this.nodes[name]; ok {
		return fmt.Errorf("RegisterFun exist:%v", name)
	}
	this.nodes[name] = node
	return this.router.Register(node, node.Route())
}

// RegisterStruct 注册一组handle
func (this *Service) RegisterStruct(i interface{}, prefix ...string) error {
	v := ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return errors.New("RegisterStruct handle type must be reflect.Struct")
	}
	if v.Elem().Kind() != reflect.Struct {
		return errors.New("RegisterStruct handle type must be reflect.Struct")
	}
	handleType := v.Type()
	serviceName := handleType.Elem().Name()

	nb := &Node{name: serviceName, binder: v, Service: this}
	if !this.filter(nb) {
		fmt.Printf("RegisterStruct filter refuse :%v,PkgPath:%v", serviceName, handleType.PkgPath)
		return nil
	}

	for m := 0; m < handleType.NumMethod(); m++ {
		method := handleType.Method(m)
		methodName := method.Name
		// value must be exported.
		if method.PkgPath != "" {
			fmt.Printf("Watch value PkgPath Not End,value:%v.%v(),PkgPath:%v", serviceName, methodName, method.PkgPath)
			continue
		}
		if !IsExported(methodName) {
			fmt.Printf("Watch value Can't Exported,value:%v.%v()", serviceName, methodName)
			continue
		}
		name := this.format(serviceName, methodName, prefix...)

		node := &Node{name: name, binder: v, value: method.Func, Service: this}
		if !this.filter(node) {
			continue
		}
		this.nodes[name] = node
		if err := this.router.Register(node, node.Route()); err != nil {
			return err
		}
	}
	return nil
}

func (this *Service) Paths() (r []string) {
	for k, _ := range this.nodes {
		r = append(r, k)
	}
	return
}

func (this *Service) Range(f func(*Node) bool) {
	for _, node := range this.nodes {
		if !f(node) {
			return
		}
	}
}
