package registry

import (
	"fmt"
	"strings"
)

const (
	PathMatchParam string = ":"
	PathMatchVague string = "*"
)

const (
	NodeTypeStatic = iota // 静态节点
	NodeTypeParam         // 参数节点
	NodeTypeWild          // 通配符节点
)

// RadixNode 基数树节点
type RadixNode struct {
	prefix   string                // 当前节点的路径前缀
	children map[string]*RadixNode // 子节点映射
	method   map[string]*Node      // HTTP方法到处理节点的映射
	nodeType int                   // 节点类型：NodeTypeStatic, NodeTypeParam, NodeTypeWild
}

// NewRadixNode 创建一个新的基数树节点
func NewRadixNode(prefix string) *RadixNode {
	return &RadixNode{
		prefix:   prefix,
		children: make(map[string]*RadixNode),
		method:   make(map[string]*Node),
		nodeType: NodeTypeStatic, // 默认是静态节点
	}
}

// Register 注册路由到基数树
func (r *RadixNode) Register(path string, method string, node *Node) {
	// 路径分割
	parts := strings.Split(path, "/")
	parts = parts[1:] // 移除空的第一个元素

	// 递归插入
	r.insert(parts, method, node)
}

// insert 递归插入路由到基数树
func (r *RadixNode) insert(parts []string, method string, node *Node) {
	if len(parts) == 0 {
		// 到达路径末尾，注册方法
		r.method[method] = node
		return
	}

	part := parts[0]
	child, exists := r.children[part]

	if !exists {
		// 检查是否为参数或通配符
		nodeType := NodeTypeStatic
		if strings.HasPrefix(part, ":") {
			nodeType = NodeTypeParam
		} else if strings.HasPrefix(part, "*") {
			nodeType = NodeTypeWild
		}

		// 创建新节点
		child = NewRadixNode(part)
		child.nodeType = nodeType
		r.children[part] = child
	}

	// 递归插入剩余部分
	child.insert(parts[1:], method, node)
}

// Match 匹配路由
func (r *RadixNode) Match(path string, method string) (*Node, map[string]string) {
	// 路径分割优化：减少内存分配
	var parts []string
	var start int
	for i, c := range path {
		if c == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}

	// 递归匹配
	params := make(map[string]string)
	node := r.match(parts, params)

	if node != nil {
		// 检查方法是否匹配
		if n, ok := node.method[method]; ok {
			return n, params
		}
	}

	return nil, nil
}

// match 递归匹配路由
func (r *RadixNode) match(parts []string, params map[string]string) *RadixNode {
	if len(parts) == 0 {
		return r
	}

	part := parts[0]

	// 优先匹配静态路由
	if child, exists := r.children[part]; exists {
		if result := child.match(parts[1:], params); result != nil {
			return result
		}
	}

	// 匹配参数路由
	for key, child := range r.children {
		if child.nodeType == NodeTypeParam {
			// 提取参数值
			paramName := strings.TrimPrefix(key, ":")
			params[paramName] = part

			if result := child.match(parts[1:], params); result != nil {
				return result
			}

			// 回溯参数
			delete(params, paramName)
		}
	}

	// 匹配通配符路由
	for _, child := range r.children {
		if child.nodeType == NodeTypeWild {
			return child
		}
	}

	return nil
}

//var RouterPrefix = []string{"/"}

func NewRouter() *Router {
	return &Router{
		static: make(map[string]*Static),
		radix:  NewRadixNode(""), // 初始化基数树
	}
}

type Static struct {
	method map[string]*Node
}

type Router struct {
	static map[string]*Static // 静态路由,不含糊任何匹配参数
	radix  *RadixNode         // 基数树路由
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

func (this *Router) Static(method string, paths ...string) (route string, node *Node) {
	route = Route(paths...)
	static, ok := this.static[route]
	if !ok {
		return
	}
	node = static.method[method]
	return
}
func (this *Router) Search(method string, paths ...string) (node *Node, params map[string]string) {
	var route string
	// 1. 最高优先级：静态路由查找
	if route, node = this.Static(method, paths...); node != nil {
		return node, nil
	}

	// 2. 基数树匹配（非静态路径）
	if this.radix != nil {
		n, p := this.radix.Match(route, method)
		if n != nil {
			return n, p
		}
	}

	return nil, nil
}

// Register 注册协议
func (this *Router) Register(node *Node, method []string) (err error) {
	if len(method) == 0 {
		return fmt.Errorf("route register method empty:%s", node.Name())
	}
	route := node.Name()
	//静态路径
	if !strings.Contains(route, PathMatchParam) && !strings.Contains(route, PathMatchVague) {
		// 初始化静态路由（如果不存在）
		if _, ok := this.static[route]; !ok {
			this.static[route] = &Static{
				method: make(map[string]*Node),
			}
		}
		// 注册方法
		for _, v := range method {
			v = strings.ToUpper(v)
			if _, ok := this.static[route].method[v]; !ok {
				this.static[route].method[v] = node
			} else {
				return fmt.Errorf("route exist:%s/%s", v, node.Name())
			}
		}
		// 静态路径不需要注册到基数树
		return nil
	}

	// 注册到基数树（非静态路径）
	for _, m := range method {
		m = strings.ToUpper(m)
		this.radix.Register(route, m, node)
	}

	return
}
