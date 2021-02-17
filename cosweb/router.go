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
	step    int              //层级
	name    string           // string,:,*,当前层匹配规则
	child   map[string]*Node //子路径
	Route   []string         //当前路由绝对路径
	Handler HandlerFunc      //handler入口
	//middleware []MiddlewareFunc //中间件
}

// Router is the registry of all registered Routes for an `Server` instance for
// Request matching and URL path parameter parsing.
type Router struct {
	root map[string]*Node //method->Node
}

func NewNode(step int, name string) *Node {
	return &Node{
		step:  step,
		name:  name,
		child: make(map[string]*Node),
	}
}

func NewRouter() *Router {
	r := &Router{
		root: make(map[string]*Node),
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
	arr := strings.Split(method+path, "/")
	lastPathIndex := len(arr) - 1

	nextNode := r.root[method]
	var spareNode []*Node

	for nextNode != nil || len(spareNode) > 0 {
		if nextNode == nil {
			ni := len(spareNode) - 1
			nextNode = spareNode[ni]
			spareNode = spareNode[0:ni]
		} else {
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
		}
		if nextNode != nil {
			if nextNode.name == RoutePathName_Vague {
				break //模糊匹配
			} else if nextNode.step == lastPathIndex && len(nextNode.Route) > 0 {
				break //精确匹配
			} else if nextNode.step >= lastPathIndex {
				nextNode = nil
			}
		}
	}
	return nextNode
}

func (r *Router) Register(route string, handler HandlerFunc, method ...string) {
	if len(method) == 0 || route == "" {
		panic("Router.Register method or route empty")
	}
	if !strings.HasPrefix(route, "/") {
		route = "/" + route
	}
	for _, m := range method {
		m = strings.ToUpper(m)
		arr := strings.Split(m+route, "/")
		if r.root[m] == nil {
			r.root[m] = NewNode(0, m)
		}
		r.root[m].addChild(arr, handler)
	}

}

func (this *Node) addChild(arr []string, handler HandlerFunc) {
	var name string
	step := this.step + 1
	length := step + 1
	if strings.HasPrefix(arr[step], RoutePathName_Param) {
		name = RoutePathName_Param
	} else if strings.HasPrefix(arr[step], RoutePathName_Vague) {
		name = RoutePathName_Vague
	} else {
		name = arr[step]
	}
	//(*)必须放在结尾
	if name == RoutePathName_Vague && len(arr) != length {
		panic(fmt.Sprintf("Router(*) must be at the end:%v", strings.Join(arr, "/")))
	}

	childNode := this.child[name]

	if childNode == nil {
		childNode = NewNode(step, name)
		this.child[name] = childNode
	} else if len(childNode.Route) > 0 {
		panic(fmt.Sprintf("Router conflict,%v == %v", childNode.String(), strings.Join(arr, "/")))
	}

	if length == len(arr) {
		childNode.Route = arr
		childNode.Handler = handler
		//childNode.middleware = middleware
	} else {
		childNode.addChild(arr, handler)
	}
}

func (this *Node) Params(path string) map[string]string {
	r := make(map[string]string)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	arr := strings.Split(path, "/")

	m := len(arr)
	if m > len(this.Route) {
		m = len(this.Route)
	}
	for i := 0; i < m; i++ {
		s := this.Route[i]
		if strings.HasPrefix(s, RoutePathName_Param) {
			k := strings.TrimPrefix(s, RoutePathName_Param)
			r[k] = arr[i]
		} else if strings.HasPrefix(s, RoutePathName_Vague) {
			k := strings.TrimPrefix(s, RoutePathName_Vague)
			if k != "" {
				r[k] = strings.Join(arr[i:], "/")
			}
		}

	}
	return r
}

func (this *Node) String() string {
	return strings.Join(this.Route, "/")
}
