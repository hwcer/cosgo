package registry

import "reflect"

type Node struct {
	name    string
	value   reflect.Value
	binder  reflect.Value //绑定的对象，作为对象的方法时才有值
	Service *Service
}

func (this *Node) Call(args ...interface{}) (r []reflect.Value) {
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

func (this *Node) Route() string {
	return Join(this.Service.prefix, this.name)
}

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

//func (this *Node) Service() *Service {
//	return this.service
//}
//
//func (this *Node) Handler() interface{} {
//	return this.service.Handler
//}

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
