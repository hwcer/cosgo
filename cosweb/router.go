package cosweb

import (
	"fmt"
	"strings"
)

const (
	RoutePathName_Param string = ":"
	RoutePathName_Vague string = "*"
)

type Node struct {
	step       int              //层级
	name       string           // string,:,*,当前层匹配规则
	child      map[string]*Node //子路径
	Route      []string         //当前路由绝对路径
	Handler    HandlerFunc      //handler入口
	Middleware []MiddlewareFunc //中间件
}

// Router is the registry of all registered Routes for an `Engine` instance for
// Request matching and URL path parameter parsing.
type Router struct {
	root   map[string]*Node //method->Node
	engine *Engine
}

func NewNode(p *Node, name string, route ...string) *Node {
	var step int
	if p != nil {
		step = p.step + 1
	}
	return &Node{
		step:  step,
		name:  name,
		Route: route,
		child: make(map[string]*Node),
	}
}

func NewRouter(e *Engine) *Router {
	r := &Router{
		root:   make(map[string]*Node),
		engine: e,
	}
	return r
}

/*
  /s/:id/update
	/s/123
*/
func (r *Router) Match(method, path string) *Node {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	method = strings.ToUpper(method)
	nextNode := r.root[method]
	if path == "/" {
		return nextNode //匹配（/） 根目录
	}

	var spareNode []*Node
	arr := strings.Split(path, "/")
	for nextNode != nil {
		i := nextNode.step + 1
		k := arr[i]
		if node := nextNode.child[RoutePathName_Vague]; node != nil {
			spareNode = append(spareNode, node)
		}
		if node := nextNode.child[RoutePathName_Param]; node != nil {
			spareNode = append(spareNode, node)
		}
		if node := nextNode.child[k]; node != nil {
			nextNode = node
		} else {
			nextNode = nil
		}

		if nextNode == nil && len(spareNode) > 0 {
			ni := len(spareNode) - 1
			nextNode = spareNode[ni]
			spareNode = spareNode[0:ni]
		}
		if nextNode != nil && len(nextNode.Route) > len(arr) {
			nextNode = nil
		}
		//模糊匹配 || 精确匹配
		if nextNode != nil && (strings.HasPrefix(nextNode.Route[nextNode.step], RoutePathName_Vague) || len(nextNode.Route) == len(arr)) {
			break
		}
	}

	return nextNode
}

func (r *Router) Register(method []string, route string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	arr := strings.Split(route, "/")
	for _, m := range method {
		m = strings.ToUpper(m)
		if r.root[m] == nil {
			r.root[m] = NewNode(nil, "")
		}
		if route != "/" {
			r.root[m].addChild(m, arr, handler, middleware...)
		}
	}
}

func (this *Node) addChild(method string, arr []string, handler HandlerFunc, middleware ...MiddlewareFunc) {
	var (
		key, name string
	)
	i := this.step + 1
	j := i + 1
	if strings.HasPrefix(arr[0], RoutePathName_Param) {
		key = strings.TrimPrefix(arr[i], RoutePathName_Param)
		name = RoutePathName_Param
	} else if strings.HasPrefix(arr[i], RoutePathName_Vague) {
		key = strings.TrimPrefix(arr[i], RoutePathName_Vague)
		name = RoutePathName_Vague
	} else {
		name = arr[0]
	}
	//(*)必须放在结尾
	if name == RoutePathName_Vague && j != len(arr) {
		panic(fmt.Sprintf("router(*) must be at the end:%v", strings.Join(arr, "/")))
	}

	childNode := this.child[name]
	//最后一层不能被覆盖
	if j == len(arr) && childNode != nil {
		panic(fmt.Sprintf("router conflict,%v:%v", method, strings.Join(arr, "/")))
	}

	if childNode == nil {
		childNode = NewNode(this, key, name, strings.Join(arr[0:j], "/"))
		this.child[name] = childNode
	}

	if j == len(arr) {
		childNode.Handler = handler
		childNode.Middleware = middleware
	} else {
		childNode.addChild(method, arr, handler, middleware...)
	}
}

func (this *Node) Params(paths []string) map[string]string {
	r := make(map[string]string)
	m := len(paths)
	if m > len(this.Route) {
		m = len(this.Route)
	}
	for i := 0; i < m; i++ {
		s := this.Route[i]
		if strings.HasPrefix(s, RoutePathName_Param) {
			k := strings.TrimPrefix(s, RoutePathName_Param)
			r[k] = paths[i]
		} else if strings.HasPrefix(s, RoutePathName_Vague) {
			k := strings.TrimPrefix(s, RoutePathName_Vague)
			if k != "" {
				r[k] = strings.Join(paths[i:], "/")
			}
		}

	}
	return r
}

func (this *Node) String() string {
	if len(this.Route) < 2 {
		return "/"
	} else {
		return strings.Join(this.Route, "/")
	}
}
