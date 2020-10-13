package express

import (
	"strconv"
	"strings"
)

type RouteMethod []string

type Route struct {
	path string

	format         bool
	prefix         string
	suffix         string
	matching       []string
	suffixMatchAll bool //匹配规则是否以 * 结束

	method     RouteMethod
	handler    HandlerFunc
	middleware []MiddlewareFunc
}

type RouteMatch struct {
	path     string
	prefix   string
	suffix   string
	matching []string
}

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
		if m == s || m == httpMethodAny {
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

//追加Method并返回更新后的Method
func (r *Route) Method(m ...string) {
	r.method = append(r.method, m...)
}

//预先格式化路径
func (r *Route) Format() {
	if r.format {
		return
	}
	//解析路径
	arr := strings.Split(r.path, "/")
	var prefix, suffix, matching []string
	//prefix
	for _, s := range arr {
		if strings.Contains(s, ":") || strings.Contains(s, "*") || len(matching) > 0 {
			matching = append(matching, s)
		} else {
			prefix = append(prefix, s)
		}
	}
	//suffix
	for i := len(matching) - 1; i >= 0; i-- {
		s := matching[i]
		if strings.Contains(s, ":") || strings.Contains(s, "*") {
			break
		} else {
			suffix = append([]string{s}, suffix...)
		}
	}
	if len(suffix) > 0 {
		matching = matching[0 : len(matching)-len(suffix)+1]
	} else if len(matching) > 0 && matching[len(matching)-1] == "*" {
		r.suffixMatchAll = true
	}

	prefix = append(prefix, "")
	if len(suffix) > 0 {
		suffix = append([]string{""}, suffix...)
	}
	r.prefix = strings.Join(prefix, "/")
	r.suffix = strings.Join(suffix, "/")
	r.matching = matching

	r.format = true

}

func (r *Route) Find(method string, path string) (param map[string]string, ok bool) {
	param = make(map[string]string)
	if !r.method.IndexOf(method) {
		return nil, false
	}
	r.Format()

	if len(r.matching) == 0 && r.path == path {
		return param, true
	}

	if r.prefix != "" {
		if strings.HasPrefix(path, r.prefix) {
			path = strings.TrimPrefix(path, r.prefix)
		} else {
			return nil, false
		}
	}
	if r.suffix != "" {
		if strings.HasPrefix(path, r.suffix) {
			path = strings.TrimSuffix(path, r.suffix)
		} else {
			return nil, false
		}
	}

	arrPath := strings.Split(path, "/")

	if len(arrPath) < len(r.matching) || (!r.suffixMatchAll && len(arrPath) > len(r.matching)) {
		return nil, false
	}

	max := len(r.matching) - 1
	if r.suffixMatchAll {
		max -= 1
	}

	var k int
	for i := 0; i <= max; i++ {
		if r.matching[i] == "*" {
			param[strconv.Itoa(k)] = arrPath[i]
			k++
			continue
		} else if strings.HasPrefix(r.matching[i], ":") {
			k := strings.TrimLeft(r.matching[i], ":")
			param[k] = arrPath[i]
		} else if r.matching[i] != arrPath[i] {
			return nil, false
		}
	}
	if r.suffixMatchAll {
		s := len(r.matching) - 1
		param[strconv.Itoa(k)] = strings.Join(arrPath[s:], "/")
	}
	return param, true
}

// SetAddress registers a new route for method and path with matching handler.
func (r *Router) Add(method, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	route := NewRoute(path, []string{method}, handler, middleware...)
	r.route = append(r.route, route)
	return route
}

func (r *Router) Match(methods []string, path string, handler HandlerFunc, middleware ...MiddlewareFunc) *Route {
	route := NewRoute(path, methods, handler, middleware...)
	r.route = append(r.route, route)
	return route
}
