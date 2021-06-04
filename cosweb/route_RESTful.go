package cosweb

import (
	"github.com/hwcer/cosgo/logger"
	"net/http"
	"reflect"
	"strings"
)

const iRESTfulRoutePath = "_RESTfulRoutePath"

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
func (this *RESTful) Route(s *Server, prefix string, method ...string) {
	arr := []string{strings.TrimSuffix(prefix, "/"), ":" + iRESTfulRoutePath}
	route := strings.Join(arr, "/")
	if len(method) == 0 {
		method = RESTfulMethods
	}
	s.Register(route, this.handle, method...)
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
		return
	}
	this.nodes[name] = handle
}

//Handle 路由入口
func (this *RESTful) handle(c *Context, next func()) (err error) {
	name := c.Get(iRESTfulRoutePath, RequestDataTypeParam)
	handle := this.nodes[name]
	if handle == nil {
		next()
		return
	}
	switch c.Request.Method {
	case http.MethodGet:
		handle.GET(c)
	case http.MethodPost:
		handle.POST(c)
	case http.MethodPut:
		handle.PUT(c)
	case http.MethodDelete:
		handle.DELETE(c)
	default:
		next()
	}
	return
}
