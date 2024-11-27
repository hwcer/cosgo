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
func newStatic(arr []string, handle interface{}) *Router {
	l := len(arr) - 1
	router := newRouter(RouteName(arr[l]), arr, l)
	router.handle = handle
	return router
}

type Router struct {
	step   int                //层级
	name   string             // string,:,*,当前层匹配规则
	route  []string           //当前路由绝对路径
	child  map[string]*Router //子路径
	static map[string]*Router //静态路由,不含糊任何匹配参数
	handle interface{}        //handle入口
	//middleware []MiddlewareFunc //中间件
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
*/
func (this *Router) Match(paths ...string) (nodes []*Router) {
	route := Join(paths...)
	//静态路由
	if v, ok := this.static[route]; ok {
		nodes = append(nodes, v)
		return
	}
	//模糊匹配
	arr := strings.Split(route, "/")
	lastPathIndex := len(arr) - 1
	if lastPathIndex == 0 {
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
	//
	//for _, node := range spareNode {
	//	fmt.Printf("spareNode:%v\n", strings.Join(node.Route(), "/"))
	//}
	n := len(spareNode)
	for selectNode != nil || n > 0 {
		if selectNode == nil {
			n -= 1
			selectNode = spareNode[n]
			spareNode = spareNode[0:n]
		}
		if selectNode.name == PathMatchVague || selectNode.step == lastPathIndex {
			if selectNode.handle != nil {
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

// Register 注册协议
func (this *Router) Register(handle interface{}, paths ...string) (err error) {
	route := Join(paths...)
	//if route == "" {
	//	return errors.New("Router.Watch method or route empty")
	//}
	arr := strings.Split(route, "/")
	//静态路径
	if !strings.Contains(route, PathMatchParam) && !strings.Contains(route, PathMatchVague) {
		if _, ok := this.static[route]; ok {
			err = fmt.Errorf("route exist:%v", route)
		} else {
			this.static[route] = newStatic(arr, handle)
		}
		return
	}
	//匹配路由
	node := this
	for i := 1; i < len(arr); i++ {
		node, err = node.addChild(arr, i)
		if err != nil {
			return
		}
	}
	if node != nil {
		node.handle = handle
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
	for i := 2; i < m; i++ {
		s := this.route[i]
		if strings.HasPrefix(s, PathMatchParam) {
			k := strings.TrimPrefix(s, PathMatchParam)
			r[k] = arr[i]
		} else if strings.HasPrefix(s, PathMatchVague) {
			k := strings.TrimPrefix(s, PathMatchVague)
			if k == "" {
				k = PathMatchVague
			}
			r[k] = strings.Join(arr[i:], "/")
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
	if node != nil && node.handle != nil && index == step {
		err = fmt.Errorf("route exist:%v", strings.Join(arr, "/"))
		return
	}
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
