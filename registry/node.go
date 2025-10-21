package registry

import (
	"reflect"
	"strings"
)

type Node struct {
	name    string //包含完整路径 /serviceName/MethodName....
	route   []string
	value   reflect.Value
	binder  reflect.Value //绑定的对象，作为对象的方法时才有值
	service *Service
}

func (this *Node) Call(args ...any) (r []reflect.Value) {
	var arr []reflect.Value
	if this.binder.IsValid() {
		arr = append(arr, this.binder)
	}
	for _, v := range args {
		arr = append(arr, reflect.ValueOf(v))
	}
	return this.value.Call(arr)
}

func (this *Node) Name() string {
	return this.name
}
func (this *Node) Route() []string {
	if this.route == nil {
		this.route = strings.Split(this.name, "/")
	}
	return this.route
}

//func (this *Node) Route() string {
//	return Join(this.service.name, this.name)
//}

func (this *Node) Value() reflect.Value {
	return this.value
}

func (this *Node) Binder() interface{} {
	if this.binder.IsValid() {
		return this.binder.Interface()
	}
	return nil
}

func (this *Node) Method() (fun interface{}) {
	return this.value.Interface()
}

func (this *Node) Service() *Service {
	return this.service
}

func (this *Node) Handler() Handler {
	return this.service.handler
}

// IsFunc 判断是func
func (this *Node) IsFunc() bool {
	return this.value.IsValid() && !this.binder.IsValid()
}

// IsMethod 对象中的方法
func (this *Node) IsMethod() bool {
	return this.value.IsValid() && this.binder.IsValid()
}

// IsStruct 仅仅是一个还未解析的对象
func (this *Node) IsStruct() bool {
	return !this.value.IsValid() && this.binder.IsValid()
}

func (this *Node) Params(paths ...string) map[string]string {
	r := make(map[string]string)
	arr := strings.Split(Join(paths...), "/")
	route := this.Route()
	m := len(arr)
	if m > len(route) {
		m = len(route)
	}
	for i := 1; i < m; i++ {
		s := route[i]
		if strings.HasPrefix(s, PathMatchParam) {
			k := strings.TrimPrefix(s, PathMatchParam)
			r[k] = arr[i]
		} else if strings.HasPrefix(s, PathMatchVague) {
			if k := strings.TrimPrefix(s, PathMatchVague); k != "" {
				r[k] = strings.Join(arr[i:], "/")
			}
		}
	}
	return r
}
