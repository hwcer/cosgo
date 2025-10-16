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
	return newRouter("", 0)
}

func newRouter(name string, step int) *Router {
	node := &Router{
		step:  step,
		name:  name,
		child: make(map[string]*Router),
		//route:  arr[0 : step+1],
		static: make(map[string]*Router),
	}
	return node
}

// 静态路由
func newStatic(arr []string) *Router {
	l := len(arr) - 1
	router := newRouter(RouteName(arr[l]), l)
	return router
}

type Router struct {
	step int    //层级
	name string // string,:,*,当前层匹配规则
	//route  []string           //当前路由绝对路径
	child  map[string]*Router //子路径
	static map[string]*Router //静态路由,不含糊任何匹配参数
	//handle *Node              //默认handle入口
	method map[string]*Node //允许的协议 短连接：GET,POST,,  [POST]:handle
	//middleware []MiddlewareFunc //中间件
}

//func (this *Router) Clone(handle *Node, route []string) *Router {
//	r := *this
//	r.child = nil
//	r.static = nil
//	r.route = route
//	r.handle = handle
//	r.method = nil
//	return &r
//}

//func (this *Router) Method() Method {
//	return Method{Router: this}
//}

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

func (this *Router) Static(method string, paths ...string) (route string, node *Node) {
	route = Route(paths...)
	r, ok := this.static[route]
	if !ok {
		return
	}
	node = r.method[method]
	return
}
func (this *Router) Search(method string, paths ...string) (nodes []*Node) {
	//静态路由
	route, node := this.Static(method, paths...)
	if node != nil {
		return []*Node{node}
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
		if r := this.child[k]; r != nil {
			spareNode = append(spareNode, r)
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
			if node = selectNode.method[method]; node != nil {
				nodes = append(nodes, node)
			}
			selectNode = nil
		} else {
			//查询子节点
			i := selectNode.step + 1
			k := arr[i]
			if r := selectNode.child[PathMatchVague]; r != nil {
				n += 1
				spareNode = append(spareNode, r)
			}
			if r := selectNode.child[PathMatchParam]; r != nil {
				n += 1
				spareNode = append(spareNode, r)
			}
			if r := selectNode.child[k]; r != nil {
				selectNode = r
			} else {
				selectNode = nil
			}
		}
	}
	return
}

//func (this *Router) Register(handle any, paths ...string) (err error) {
//	route := Route(paths...)
//	return this.register(handle, route)
//}

func (this *Router) setRouterNode(router *Router, node *Node, method []string) (err error) {
	if router.method == nil {
		router.method = make(map[string]*Node)
	}
	for _, v := range method {
		v = strings.ToUpper(v)
		if _, ok := router.method[v]; !ok {
			router.method[v] = node
		} else {
			return fmt.Errorf("route exist:%s/%s", v, node.Name())
		}
	}
	return nil
}

// Register 注册协议
func (this *Router) Register(node *Node, method []string) (err error) {
	if len(method) == 0 {
		return fmt.Errorf("route register method empty:%s", node.Name())
	}
	route := node.Name()
	arr := strings.Split(route, "/")
	//静态路径
	if !strings.Contains(route, PathMatchParam) && !strings.Contains(route, PathMatchVague) {
		router := newStatic(arr)
		this.static[route] = router
		return this.setRouterNode(router, node, method)
	}
	//匹配路由

	router := this
	for i := 1; i < len(arr); i++ {
		if router, err = router.addChild(arr, i); err != nil {
			return
		}
	}
	if router != nil {
		err = this.setRouterNode(router, node, method)
	}
	return
}

//func (this *Router) Route() (r []string) {
//	r = append(r, this.route...)
//	return r
//}

//func (this *Router) Handle() interface{} {
//	return this.handle
//}

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
		node = newRouter(name, step)
		this.child[name] = node
	}
	return
}

//func (this *Router) String() string {
//	if len(this.route) < 2 {
//		return "/"
//	}
//	return strings.Join(this.route, "/")
//}
//
//func (this *Router) childes() (r []string) {
//	for _, c := range this.child {
//		r = append(r, c.String())
//	}
//	return
//}
