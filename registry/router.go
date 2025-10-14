package registry

import (
	"fmt"
	"runtime/debug"
	"strings"
)

const (
	PathMatchParam string = ":"
	PathMatchVague string = "*"
)

//var RouterPrefix = []string{"/"}

func NewRouter() *Router {
	return newRouter("", []string{""}, 0)
}

func newRouter(name string, arr []string, step int) *Router {
	node := &Router{
		step:   step,
		name:   name,
		child:  make(map[string]*Router),
		route:  arr[0 : step+1],
		static: make(map[string]*Router),
	}
	return node
}

// 静态路由
func newStatic(arr []string) *Router {
	l := len(arr) - 1
	router := newRouter(RouteName(arr[l]), arr, l)
	return router
}

type Router struct {
	step   int                //层级
	name   string             // string,:,*,当前层匹配规则
	route  []string           //当前路由绝对路径
	child  map[string]*Router //子路径
	static map[string]*Router //静态路由,不含糊任何匹配参数
	handle any                //默认handle入口
	method map[string]*Router //允许的协议 短连接：GET,POST,,  [POST]:handle
	//middleware []MiddlewareFunc //中间件
}

func (this *Router) Clone(handle any, route []string) *Router {
	r := *this
	r.child = nil
	r.static = nil
	r.route = route
	r.handle = handle
	r.method = nil
	return &r
}

func (this *Router) Method() Method {
	return Method{Router: this}
}

// Router is the registry of all registered Routes for an `engine` instance for
// Request matching and URL path parameter parsing.
//type Router struct {
//	root   map[string]*Router //method->Router
//	static map[string]*Router //静态路由,不含糊任何匹配参数
//}

/*
/s/:id/update
/s/123
/GET/*
*/

func (this *Router) Match(paths ...string) (nodes []*Router) {
	route := Route(paths...)
	//静态路由
	if v, ok := this.static[route]; ok {
		nodes = append(nodes, v)
		return
	}
	arr := strings.Split(route, "/")
	//模糊匹配
	lastPathIndex := len(arr) - 1
	if lastPathIndex < 1 {
		fmt.Printf("错误的路由地址:%v\n%v", route, string(debug.Stack()))
		return
	}

	var spareNode []*Router
	var selectNode *Router

	for _, k := range []string{PathMatchVague, PathMatchParam, arr[1]} {
		if node := this.child[k]; node != nil {
			spareNode = append(spareNode, node)
		}
	}

	n := len(spareNode)
	for selectNode != nil || n > 0 {
		if selectNode == nil {
			n -= 1
			selectNode = spareNode[n]
			spareNode = spareNode[0:n]
		}
		if selectNode.name == PathMatchVague || selectNode.step == lastPathIndex {
			if selectNode.handle != nil || len(selectNode.method) > 0 {
				nodes = append(nodes, selectNode)
			}
			selectNode = nil
		} else {
			//查询子节点
			i := selectNode.step + 1
			k := arr[i]
			if node := selectNode.child[PathMatchVague]; node != nil {
				n += 1
				spareNode = append(spareNode, node)
			}
			if node := selectNode.child[PathMatchParam]; node != nil {
				n += 1
				spareNode = append(spareNode, node)
			}
			if node := selectNode.child[k]; node != nil {
				selectNode = node
			} else {
				selectNode = nil
			}
		}
	}
	return
}

func (this *Router) Register(handle any, paths ...string) (err error) {
	route := Route(paths...)
	return this.register(handle, route)
}

func (this *Router) setRouterHandle(node *Router, handle any, route []string, method ...string) (err error) {
	if len(method) == 0 {
		if node.handle == nil {
			node.handle = handle
		} else {
			err = fmt.Errorf("route exist:%s", node.name)
		}
		return
	}
	if node.method == nil {
		node.method = make(map[string]*Router)
	}
	for _, v := range method {
		if _, ok := node.method[v]; !ok {
			node.method[v] = node.Clone(handle, route)
		} else {
			return fmt.Errorf("route exist:%s/%s", v, node.name)
		}
	}
	return nil
}

// Register 注册协议
func (this *Router) register(handle any, route string, method ...string) (err error) {
	arr := strings.Split(route, "/")
	//静态路径
	if !strings.Contains(route, PathMatchParam) && !strings.Contains(route, PathMatchVague) {
		node := newStatic(arr)
		this.static[route] = node
		return this.setRouterHandle(node, handle, arr, method...)
	}
	//匹配路由
	node := this
	for i := 1; i < len(arr); i++ {
		if node, err = node.addChild(arr, i); err != nil {
			return
		}
	}
	if node != nil {
		err = this.setRouterHandle(node, handle, arr, method...)
	}
	return
}

func (this *Router) Route() (r []string) {
	r = append(r, this.route...)
	return r
}

func (this *Router) Handle() interface{} {
	return this.handle
}

func (this *Router) Params(paths ...string) map[string]string {
	r := make(map[string]string)
	arr := strings.Split(Join(paths...), "/")
	m := len(arr)
	if m > len(this.route) {
		m = len(this.route)
	}
	for i := 1; i < m; i++ {
		s := this.route[i]
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

func (this *Router) addChild(arr []string, step int) (node *Router, err error) {
	name := RouteName(arr[step])
	index := len(arr) - 1
	//(*)必须放在结尾
	if name == PathMatchVague && index != step {
		return nil, fmt.Errorf("router(*) must be at the end:%v", strings.Join(arr, "/"))
	}
	//路由重复
	node = this.child[name]
	if node == nil {
		node = newRouter(name, arr, step)
		this.child[name] = node
	}
	return
}

func (this *Router) String() string {
	if len(this.route) < 2 {
		return "/"
	}
	return strings.Join(this.route, "/")
}

func (this *Router) childes() (r []string) {
	for _, c := range this.child {
		r = append(r, c.String())
	}
	return
}

type Method struct {
	*Router
}

func (this *Method) Match(method string, paths ...string) []*Router {
	nodes := this.Router.Match(paths...)
	var arr []*Router
	for _, node := range nodes {
		if n, ok := node.method[method]; ok {
			arr = append(arr, n)
		} else if node.handle != nil {
			arr = append(arr, node)
		}
	}
	return arr
}

func (this *Method) Register(handle any, route string, method ...string) error {
	route = Route(route)
	return this.Router.register(handle, route, method...)
}
