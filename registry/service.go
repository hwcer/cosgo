package registry

import (
	"errors"
	"fmt"
	"path"
	"reflect"
	"strings"
)

type Handler interface {
	Filter(*Node) bool
}

// NewService
// name   :  必须是使用Route格式化后的路径
func NewService(name string, router *Router, handle ...Handler) *Service {
	r := &Service{
		name:   name,
		nodes:  make(map[string]*Node),
		router: router,
	}
	if len(handle) > 0 {
		r.handler = handle[0]
	}
	return r
}

type Service struct {
	name string // a/b
	//prefix  string //  /a/b
	nodes   map[string]*Node
	router  *Router
	methods []string
	handler Handler //自定义 Filter等方法
}

//func (this *Service) Use(i any) {
//	this.Handler.Use(i)
//}

func (this *Service) Name() string {
	return this.name
}

// SetMethods 设置默认请求方式
func (this *Service) SetMethods(s []string) {
	this.methods = s
}
func (this *Service) Handler() Handler {
	return this.handler
}

//func (this *Service) Prefix() string {
//	return this.prefix
//}

//	func (this *Service) Merge(s *Service) (err error) {
//		if s.Handler != nil {
//			this.Handler = s.Handler
//		}
//		for k, v := range s.nodes {
//			node := &Node{name: v.name, value: v.value, binder: v.binder, Service: this}
//			this.nodes[k] = node
//			if err = this.router.Register(node, node.Route()); err != nil {
//				return
//			}
//		}
//		return
//	}
func (this *Service) Register(i any, prefix ...string) error {
	nodes, err := this.Parse(i, prefix...)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if err = this.router.Register(node, this.methods); err == nil {
			nodes = append(nodes, node)
		} else {
			fmt.Printf("router register error,route:%s error:%v", node.Name(), err)
		}
	}
	return nil
}

func (this *Service) RegisterWithMethod(i any, method []string, prefix ...string) error {
	nodes, err := this.Parse(i, prefix...)
	if err != nil {
		return err
	}
	for _, node := range nodes {
		if err = this.router.Register(node, method); err == nil {
			nodes = append(nodes, node)
		} else {
			fmt.Printf("router register error,route:%s error:%v", node.Name(), err)
		}
	}
	return nil
}

// Parse 解析服务
func (this *Service) Parse(i any, prefix ...string) (nodes []*Node, err error) {
	v := reflect.ValueOf(i)
	var kind reflect.Kind
	if v.Kind() == reflect.Ptr {
		kind = v.Elem().Kind()
	} else {
		kind = v.Kind()
	}
	switch kind {
	case reflect.Func:
		nodes, err = this.ParseFun(v, prefix...)
	case reflect.Struct:
		nodes, err = this.ParseStruct(v, prefix...)
	default:
		err = fmt.Errorf("未知的注册类型：%v", v.Kind())
	}
	return
}

func (this *Service) filter(node *Node) bool {
	if this.handler == nil {
		return true
	}
	return this.handler.Filter(node)
}

func (this *Service) format(serviceName, methodName string, prefix ...string) string {
	if len(prefix) == 0 {
		return Route(this.name, serviceName, methodName)
	}
	arr := append([]string{this.name}, prefix...)
	p := Route(arr...)
	var name string
	if serviceName == "" {
		name = methodName
	} else {
		name = path.Join(serviceName, methodName)
	}
	p = strings.Replace(p, "%v", Formatter(name), -1)
	p = strings.Replace(p, "%s", Formatter(serviceName), -1)
	p = strings.Replace(p, "%m", Formatter(methodName), -1)
	return p
}

// Node 创建Node
//func (this *Service) Node(i any, prefix ...string) (*Node, error) {
//	v := ValueOf(i)
//	if v.Kind() != reflect.Func {
//		return nil, errors.New("RegisterFun fn type must be reflect.Func")
//	}
//	name := this.format(this.name, FuncName(v), prefix...)
//	node := &Node{name: name, value: v}
//	if !this.filter(node) {
//		return nil, fmt.Errorf("service filter error:%v", name)
//	}
//	return node, nil
//}

func (this *Service) ParseFun(i interface{}, prefix ...string) (nodes []*Node, err error) {
	v := ValueOf(i)
	if v.Kind() != reflect.Func {
		return nil, errors.New("RegisterFun fn type must be reflect.Func")
	}
	name := this.format("", FuncName(v), prefix...)
	node := &Node{name: name, value: v, service: this}
	if !this.filter(node) {
		return nil, fmt.Errorf("service filter error:%v", name)
	}
	if _, ok := this.nodes[node.name]; ok {
		return nil, fmt.Errorf("ParseFun exist:%v", node.name)
	}
	nodes = append(nodes, node)
	this.nodes[node.name] = node

	return
}

// ParseStruct 注册一组handle
func (this *Service) ParseStruct(i interface{}, prefix ...string) ([]*Node, error) {
	v := ValueOf(i)
	if v.Kind() != reflect.Ptr {
		return nil, errors.New("RegisterStruct handle type must be reflect.Struct")
	}
	if v.Elem().Kind() != reflect.Struct {
		return nil, errors.New("RegisterStruct handle type must be reflect.Struct")
	}
	handleType := v.Type()
	serviceName := handleType.Elem().Name()

	nb := &Node{name: serviceName, binder: v, service: this}
	if !this.filter(nb) {
		fmt.Printf("RegisterStruct filter refuse :%v,PkgPath:%v", serviceName, handleType.PkgPath())
		return nil, nil
	}

	var nodes []*Node

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
		node := &Node{name: name, binder: v, value: method.Func, service: this}
		if this.filter(node) {
			nodes = append(nodes, node)
			this.nodes[name] = node
		}
	}
	return nodes, nil
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

//func (this *Service) Method() *Method {
//	return &Method{Service: this}
//}
//
//type Method struct {
//	*Service
//}
//

//
//func (this *Method) Register(i any, route string, method ...string) error {
//	route = Route(route)
//	return this.Router.register(handle, route, method...)
//}
