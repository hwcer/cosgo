package express

import (
	"fmt"
	"reflect"
)

var typeOfContext = reflect.TypeOf(&Context{})

//NameSpace  namsespace代表了一个虚拟目录，目录下有很多method
//可以使用 /path/$NameSpace/$method 的格式访问
type NameSpace struct {
	name  string
	nodes map[string]*NameSpaceMethod
}

type NameSpaceMethod struct {
	proto  reflect.Value
	value  reflect.Value
	method map[string]func(*Context) error
}

func NewNameSpace(name string) *NameSpace {
	return &NameSpace{
		name:  name,
		nodes: make(map[string]*NameSpaceMethod),
	}
}

//Register 注册一组handle，重名忽略
func (this *NameSpace) Register(handle interface{}) {
	typ := reflect.TypeOf(handle)
	name := typ.Elem().Name()
	proto := reflect.ValueOf(handle)
	node := this.nodes[name]
	if node == nil {
		node = &NameSpaceMethod{
			proto:  reflect.ValueOf(handle),
			method: make(map[string]func(*Context) error),
		}
		this.nodes[name] = node
	}
	//fmt.Printf("Register:%v\n",NameSpace.name)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		//fmt.Println("打印Method", mname, mtype)
		// method must be exported.
		if method.PkgPath != "" {
			fmt.Printf("Register method PkgPath Not End,method:%v.%v(),PkgPath:%v\n", name, mname, method.PkgPath)
			continue
		}
		if !isExported(mname) {
			fmt.Printf("Register method Can't Exported,method:%v.%v()\n", name, mname)
			continue
		}
		// method needs four ins: receiver, context.Context, *args, *reply.
		if mtype.NumIn() != 2 || mtype.NumOut() != 1 {
			fmt.Printf("Register method args num or return num error,method:%v.%v()\n", name, mname)
			continue
		}
		// First arg must be context.Context
		//ctxType := mtype.In(1)
		//if !ctxType.ConvertibleTo(typeOfContext) {
		//	fmt.Printf("Register method args error,method:%v.%v()\n",NameSpace.name,mname)
		//	continue
		//}
		////
		//outType := mtype.Out(0)
		//if !outType.ConvertibleTo(typeOfMessage) {
		//	fmt.Printf("Register method return error,method:%v.%v()\n",NameSpace.name,mname)
		//	continue
		//}

		//service.method[mname] = method.Func.Interface().(handlerMethod)
		//fmt.Printf("Register method；%v\n",mname)
		node.method[mname] = method.Func

	}
}
