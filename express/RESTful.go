package express

import (
	"cosgo/logger"
	"net/http"
	"reflect"
)

const iRESTfulRoutePath = "_RESTful_PATH"

var RESTfulMethods = []string{
	http.MethodGet,
	http.MethodPut,
	http.MethodPost,
	http.MethodDelete,
}

type iRESTful interface {
	GET(*Context) error    //用来获取资源
	PUT(*Context) error    //PUT用来更新资源
	POST(*Context) error   //用来新建资源（也可以用于更新资源）
	DELETE(*Context) error //用来删除资源
}

type RESTful struct {
	nodes map[string]iRESTful
}

func NewRESTful() *RESTful {
	return &RESTful{
		nodes: make(map[string]iRESTful),
	}
}

//Handle 路由入口
func (this *RESTful) handler(c *Context) error {
	name := c.Param(iRESTfulRoutePath)
	if name == "" {
		return nil
	}
	handle := this.nodes[name]
	if handle == nil {
		return nil
	}

	switch c.Request.Method {
	case http.MethodGet:
		return handle.GET(c)
	case http.MethodPost:
		return handle.POST(c)
	case http.MethodPut:
		return handle.PUT(c)
	case http.MethodDelete:
		return handle.DELETE(c)
	default:
		return nil
	}
}

//Register 注册一组handle，重名忽略
func (this *RESTful) Register(handle iRESTful) {
	handleType := reflect.TypeOf(handle)
	if handleType.Kind() != reflect.Ptr {
		logger.Error("RESTful Register error:handle not pointer")
		return
	}
	name := strFirstToLower(handleType.Elem().Name())
	if _, ok := this.nodes[name]; ok {
		logger.Error("RESTful Register error:%v exist", name)
	}
	this.nodes[name] = handle
}
