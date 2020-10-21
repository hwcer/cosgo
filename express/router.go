package express

import (
	"strconv"
	"strings"
)

type Route struct {
	path string

	format   bool
	prefix   string
	matching []string

	staticMatch bool //静态路由
	suffixMatch bool //匹配规则是否以 * 结束

	method     RouteMethod
	handler    HandlerFunc
	middleware []MiddlewareFunc
}

type RouteMethod []string

// Router is the registry of all registered Routes for an `Engine` instance for
// Request matching and URL path parameter parsing.
type Router struct {
	route  []*Route
	engine *Engine
}

func NewRoute(path string, method []string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	route := &Route{
		path:       path,
		method:     method,
		handler:    handler,
		middleware: middleware,
	}
	route.Format()
	return route
}

// NewRouter returns a new Router instance.
func NewRouter(e *Engine) *Router {
	return &Router{
		engine: e,
		route:  make([]*Route, 0),
	}
}

func (r RouteMethod) IndexOf(s string) bool {
	for _, m := range r {
		if m == s || m == HttpMethodAny {
			return true
		}
	}
	return false
}

//加入中间件
func (r *Route) Use(m ...MiddlewareFunc) {
	r.middleware = append(r.middleware, m...)
}

//原始路径
func (r *Route) Path() string {
	return r.path
}

//匹配路由
func (r *Route) Match(method string, path string) (param map[string]string, ok bool) {
	param = make(map[string]string)
	if !r.method.IndexOf(method) {
		return nil, false
	}
	r.Format()
	if !strings.HasPrefix(path, r.prefix) {
		return nil, false
	}
	//静态路由
	if r.staticMatch {
		if r.suffixMatch && len(path) > len(r.prefix) {
			ok = true
			param["0"] = strings.TrimPrefix(path, r.prefix)
		} else if !r.suffixMatch && path == r.prefix {
			ok = true
		}
		return
	}

	arrPath := strings.Split(strings.TrimPrefix(path, r.prefix), "/")
	//通配符尾缀
	var suffix string
	if r.suffixMatch {
		if len(arrPath) <= len(r.matching) {
			return nil, false
		}
		suffix = strings.Join(arrPath[len(r.matching):], "/")
		if suffix == "" {
			return nil, false
		}
		arrPath = arrPath[0:len(r.matching)]
	} else if len(arrPath) != len(r.matching) {
		return nil, false
	}

	var k int
	for i := 0; i < len(r.matching); i++ {
		if r.matching[i] == "*" {
			param[strconv.Itoa(k)] = arrPath[i]
			k++
		} else if strings.HasPrefix(r.matching[i], ":") {
			k := strings.TrimPrefix(r.matching[i], ":")
			param[k] = arrPath[i]
		} else if r.matching[i] != arrPath[i] {
			return nil, false
		}
	}
	if r.suffixMatch {
		param[strconv.Itoa(k)] = suffix
	}
	return param, true
}

//追加Method并返回更新后的Method
func (r *Route) Method(m ...string) {
	r.method = append(r.method, m...)
}

//预先格式化路径
func (r *Route) Format() {
	if r.format {
		return
	}
	path := r.path
	if strings.HasSuffix(path, "*") {
		path = strings.TrimSuffix(path, "*")
		r.suffixMatch = true
	}
	if !strings.Contains(path, "*") && !strings.Contains(path, ":") {
		r.prefix = path
		r.staticMatch = true
		return
	}
	if r.suffixMatch {
		path = strings.TrimSuffix(path, "/")
	}
	//解析路径
	arr := strings.Split(path, "/")
	var prefix, matching []string
	//prefix
	for _, s := range arr {
		if len(matching) > 0 || strings.Contains(s, ":") || strings.Contains(s, "*") {
			matching = append(matching, s)
		} else {
			prefix = append(prefix, s)
		}
	}

	if len(prefix) > 0 {
		prefix = append(prefix, "") //以"/"结束
	}
	r.prefix = strings.Join(prefix, "/")
	r.matching = matching

	r.format = true

}

// SetAddress registers a new route for value and path with matching handler.
func (r *Router) Route(method []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	route := NewRoute(path, method, handler, middleware...)
	r.route = append(r.route, route)
	return route
}

//通过一个对象注册，导入对象中的所有方法
func (r *Router) Register(method []string, path string, i interface{}, middleware ...MiddlewareFunc) (*Route, *NameSpace) {
	arr := []string{strings.TrimSuffix(path, "/"), ":" + nameSpacePathName, ":" + nameSpaceMethodName}
	nsp := NewNameSpace()
	nsp.Register(i)
	route := NewRoute(strings.Join(arr, "/"), method, nsp.handler, middleware...)
	r.route = append(r.route, route)
	return route, nsp
}