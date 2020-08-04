package server

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)


var typeOfMessage = reflect.TypeOf(&Message{})
var typeOfContext = reflect.TypeOf(&gin.Context{})



type Router struct {
	handlers map[string]*nsp
	PathParser func(*gin.Context) []string  //路径解析，返回[path...,method]
}



func NewRouter() *Router {
	return &Router{handlers: make(map[string]*nsp)}
}


//nsp  namsespace代表了一个虚拟目录，目录下有很多method
//可以使用 url/$nsp/$method 的格式访问
type nsp struct {
	name  string
	method map[string]*nspMethod
}

type nspMethod struct {
	proto reflect.Value
	value reflect.Value
	method func(*gin.Context)*Message
}


func newNsp(name string) *nsp {
	return &nsp{
		name:          name,
		method: make(map[string]*nspMethod),
	}
}

func (n *nsp)Register(fun func(*gin.Context)*Message)  {

}

//获得一个名字空间
func (s *Router) Nsp(autoCreate bool,args ...string) *nsp {
	name := joinPath(args...)
	nsp := s.handlers[name]
	if nsp == nil &&  autoCreate{
		nsp = newNsp(name)
		s.handlers[name] = nsp
	}
	return nsp
}


//Register 注册一组handle，重名忽略
func (s *Router)Register( handle interface{}) *nsp{
	typ := reflect.TypeOf(handle)
	nsp := s.Nsp(true,typ.Elem().Name())
	proto := reflect.ValueOf(handle)
	//fmt.Printf("Register:%v\n",nsp.name)
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := method.Name
		//fmt.Println("打印Method", mname, mtype)
		// Method must be exported.
		if method.PkgPath != ""  {
			fmt.Printf("Register Method PkgPath Not Empty,Method:%v.%v(),PkgPath:%v\n",nsp.name,mname,method.PkgPath)
			continue
		}
		if !isExported(mname) {
			fmt.Printf("Register Method Can't Exported,Method:%v.%v()\n",nsp.name,mname)
			continue
		}
		// Method needs four ins: receiver, context.Context, *args, *reply.
		if mtype.NumIn() != 2 || mtype.NumOut() !=1 {
			fmt.Printf("Register Method args num or return num error,Method:%v.%v()\n",nsp.name,mname)
			continue
		}
		// First arg must be context.Context
		//ctxType := mtype.In(1)
		//if !ctxType.ConvertibleTo(typeOfContext) {
		//	fmt.Printf("Register Method args error,Method:%v.%v()\n",nsp.name,mname)
		//	continue
		//}
		////
		//outType := mtype.Out(0)
		//if !outType.ConvertibleTo(typeOfMessage) {
		//	fmt.Printf("Register Method return error,Method:%v.%v()\n",nsp.name,mname)
		//	continue
		//}

		//service.method[mname] = method.Func.Interface().(handlerMethod)
		//fmt.Printf("Register Method；%v\n",mname)
		nsp.method[mname] = &nspMethod{
			proto:  proto,
			value:  method.Func,
			method: nil,
		}

	}
	s.handlers[nsp.name] = nsp
	//func

	return nsp
	//logger.DEBUG("RegisterRef %v  %+v", nsp, service)
}


//Handle gin路由入口
func (s *Router) Handle(c *gin.Context) {
	defer func() {
		if err := recover(); err != nil {
			c.String(http.StatusInternalServerError, fmt.Sprint(err))
		}
	}()

	var route []string
	if s.PathParser != nil{
		route = s.PathParser(c)
	}else{
		route = defPathParse(c)
	}


	path := joinPath(route[0:len(route)-1]...)
	name := strFirstToUpper(strings.Trim(route[len(route)-1],"/"))

	nsp := s.handlers[path]
	if nsp == nil {
		c.String(http.StatusNotFound, "NSP NOT EXIST：%v", path)
		return
	}
	method :=  nsp.method[name]
	if method == nil {
		c.String(http.StatusNotFound, "METHOD NOT EXIST：%v", name)
		return
	}

	var reply *Message
	if method.method!=nil {
		reply = method.method(c)
	} else if !method.value.IsNil() {
		ret := method.value.Call([]reflect.Value{method.proto, reflect.ValueOf(c)})
		reply = ret[0].Interface().(*Message)
	}else{
		c.String(http.StatusNotFound, "NSP METHOD EMPTY：%v", route)
		return
	}
	//fmt.Printf("router handle reply；%+v",reply)
	status := reply.GetHttpStatus()
	dataType := reply.GetDataType()

	if Config.ErrHeaderName !="" {
		c.Header(Config.ErrHeaderName, strconv.Itoa(reply.Code))
	}
	if reply.HasError(){
		c.String(status, reply.Error)
	}else if dataType == MessageDataType_String{
		c.String(status,reply.Data.(string))
	}else if dataType == MessageDataType_Json{
		c.JSON(status,reply.Data)
	}else if dataType == MessageDataType_Xml{
		c.XML(status,reply.Data)
	}else if dataType == MessageDataType_Protobuf{
		c.ProtoBuf(status,reply.Data)
	}else{
		c.JSON(status, reply)
	}
}










