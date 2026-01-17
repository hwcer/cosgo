package registry

import (
	"fmt"
	"strings"

	"github.com/hwcer/cosgo/slice"
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
	prefix     string                // 当前节点的路径前缀
	children   map[string]*RadixNode // 子节点映射
	method     map[string]*Node      // HTTP方法到处理节点的映射
	nodeType   int                   // 节点类型：NodeTypeStatic, NodeTypeParam, NodeTypeWild
	paramNames []string              // 参数节点的参数名列表
}

// NewRadixNode 创建一个新的基数树节点
func NewRadixNode(prefix string) *RadixNode {
	return &RadixNode{
		prefix:     prefix,
		children:   make(map[string]*RadixNode),
		method:     make(map[string]*Node),
		nodeType:   NodeTypeStatic, // 默认是静态节点
		paramNames: []string{},     // 参数名列表
	}
}

// Register 注册路由到基数树
func (r *RadixNode) Register(path string, method []string, node *Node) error {
	// 路径分割
	parts := strings.Split(path, "/")
	if parts[0] == "" {
		parts = parts[1:] // 移除空的第一个元素
	}
	// 递归插入
	return r.insert(path, parts, method, node)
}

// insert 递归插入路由到基数树
func (r *RadixNode) insert(path string, parts []string, method []string, node *Node) error {
	if len(parts) == 0 {
		// 到达路径末尾，注册方法
		for _, m := range method {
			m = strings.ToUpper(m)
			// 检查方法是否已经存在
			if _, exists := r.method[m]; exists {
				return fmt.Errorf("method already exists: %s for path: %s", m, path)
			}
			// 如果方法不存在，那么注册方法
			r.method[m] = node
		}
		return nil
	}

	part := parts[0]
	// 确定节点前缀
	var prefix string
	if strings.HasPrefix(part, PathMatchParam) {
		prefix = PathMatchParam
	} else if strings.HasPrefix(part, PathMatchVague) {
		prefix = PathMatchVague
	} else {
		prefix = Formatter(part)
	}
	child, exists := r.children[prefix]

	if !exists {
		// 检查是否为参数或通配符
		nodeType := NodeTypeStatic
		var paramName string
		if strings.HasPrefix(part, PathMatchParam) {
			nodeType = NodeTypeParam
			paramName = strings.TrimPrefix(part, PathMatchParam)
		} else if strings.HasPrefix(part, PathMatchVague) {
			nodeType = NodeTypeWild
			//paramName = PathMatchVague // 通配符参数名在匹配时固定为 "*"，不需要特意处理
		}

		// 创建新节点
		child = NewRadixNode(prefix)
		child.nodeType = nodeType
		if paramName != "" {
			child.paramNames = append(child.paramNames, paramName)
		}
		r.children[prefix] = child
	} else if child.nodeType == NodeTypeParam && strings.HasPrefix(part, PathMatchParam) {
		// 如果已经存在参数节点但参数名不同，添加新的参数名
		paramName := strings.TrimPrefix(part, PathMatchParam)
		// 检查参数名是否已经存在
		if !slice.Has(child.paramNames, paramName) {
			child.paramNames = append(child.paramNames, paramName)
		}
	}

	// 递归插入剩余部分
	return child.insert(path, parts[1:], method, node)
}

// Match 匹配路由
func (r *RadixNode) Match(path string, method string) (*Node, map[string]string) {
	// 使用公共的 Split 方法进行路径分割
	parts := Split(path)

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
	formattedPart := Formatter(part)
	if child, exists := r.children[formattedPart]; exists {
		if result := child.match(parts[1:], params); result != nil {
			return result
		}
	}

	// 匹配参数路由
	if child, exists := r.children[PathMatchParam]; exists && child.nodeType == NodeTypeParam {
		// 提取参数值，对所有参数名赋值
		for _, paramName := range child.paramNames {
			params[paramName] = part
		}

		if result := child.match(parts[1:], params); result != nil {
			return result
		}

		// 回溯参数
		for _, paramName := range child.paramNames {
			delete(params, paramName)
		}
	}

	// 匹配通配符路由
	if child, exists := r.children[PathMatchVague]; exists && child.nodeType == NodeTypeWild {
		// 提取剩余路径作为通配符参数值
		if len(parts) > 0 {
			// 构建剩余路径
			remainingPath := strings.Join(parts, "/")
			// 通配符参数名在匹配时固定为 "*"
			params[PathMatchVague] = remainingPath
		}
		return child
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
	// 1. 最高优先级：静态路由查找
	if _, node = this.Static(method, paths...); node != nil {
		return node, nil
	}

	// 2. 基数树匹配（非静态路径）
	if this.radix != nil {
		originalPath := "/" + strings.Join(paths, "/")
		n, p := this.radix.Match(originalPath, method)
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
		static, ok := this.static[route]
		if !ok {
			static = &Static{
				method: make(map[string]*Node),
			}
			this.static[route] = static
		}
		// 注册方法
		for _, v := range method {
			v = strings.ToUpper(v)
			if _, ok := static.method[v]; !ok {
				static.method[v] = node
			} else {
				return fmt.Errorf("route exist:%s/%s", v, node.Name())
			}
		}
		// 静态路径不需要注册到基数树
		return nil
	}

	// 注册到基数树（非静态路径）
	if err := this.radix.Register(route, method, node); err != nil {
		return err
	}

	return
}
