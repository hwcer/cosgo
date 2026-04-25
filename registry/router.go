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
	NodeTypeStatic = iota
	NodeTypeParam
	NodeTypeWild
)

// methodEntry 方法 → 处理节点映射
// 用切片替代 map：路由通常只注册 1~3 个 HTTP 方法，线性扫描比 hash 更快
type methodEntry struct {
	method string
	node   *Node
}

// staticChild 静态子节点条目
type staticChild struct {
	key  string
	node *RadixNode
}

// RadixNode 路由树节点
// children 拆分为 statics/paramChild/wildChild 三个字段，
// 匹配时��接按优先级（静态 > 参数 > 通配符）访问，消除 map hash 开销
type RadixNode struct {
	prefix     string
	statics    []staticChild // 静态子节点（按 key 线性扫描，典型 1~5 个）
	paramChild *RadixNode    // 参数子节点（最多一个 ":"）
	wildChild  *RadixNode    // 通配符子节点（最多一个 "*"）
	methods    []methodEntry // HTTP 方法 → Node
	nodeType   int
	paramNames []string
}

func NewRadixNode(prefix string) *RadixNode {
	return &RadixNode{prefix: prefix}
}

// findStatic 线性扫描静态子节点
func (r *RadixNode) findStatic(key string) *RadixNode {
	for i := range r.statics {
		if r.statics[i].key == key {
			return r.statics[i].node
		}
	}
	return nil
}

// findMethod 线性扫描方法
func (r *RadixNode) findMethod(method string) *Node {
	for i := range r.methods {
		if r.methods[i].method == method {
			return r.methods[i].node
		}
	}
	return nil
}

// hasMethod 是否有注册的方法
func (r *RadixNode) hasMethod() bool {
	return len(r.methods) > 0
}

// Register 注册路由到树
func (r *RadixNode) Register(path string, method []string, node *Node) error {
	parts := Split(path)
	return r.insert(path, parts, method, node)
}

func (r *RadixNode) insert(path string, parts []string, method []string, node *Node) error {
	if len(parts) == 0 {
		for _, m := range method {
			m = strings.ToUpper(m)
			if r.findMethod(m) != nil {
				return fmt.Errorf("method already exists: %s for path: %s", m, path)
			}
			r.methods = append(r.methods, methodEntry{method: m, node: node})
		}
		return nil
	}

	part := parts[0]

	if strings.HasPrefix(part, PathMatchParam) {
		paramName := strings.TrimPrefix(part, PathMatchParam)
		if r.paramChild == nil {
			r.paramChild = NewRadixNode(PathMatchParam)
			r.paramChild.nodeType = NodeTypeParam
		}
		if paramName != "" && !slice.Has(r.paramChild.paramNames, paramName) {
			r.paramChild.paramNames = append(r.paramChild.paramNames, paramName)
		}
		return r.paramChild.insert(path, parts[1:], method, node)
	}

	if strings.HasPrefix(part, PathMatchVague) {
		if r.wildChild == nil {
			r.wildChild = NewRadixNode(PathMatchVague)
			r.wildChild.nodeType = NodeTypeWild
		}
		return r.wildChild.insert(path, parts[1:], method, node)
	}

	key := Formatter(part)
	child := r.findStatic(key)
	if child == nil {
		child = NewRadixNode(key)
		r.statics = append(r.statics, staticChild{key: key, node: child})
	}
	return child.insert(path, parts[1:], method, node)
}

// Match 匹配路由
// 零 Split：直接扫描 path 提取段
// Params 用切片：回溯仅需 reslice（O(1)），无 map delete
func (r *RadixNode) Match(path string, method string) (*Node, Params) {
	var params Params
	pos := 0
	if len(path) > 0 && path[0] == '/' {
		pos = 1
	}
	node := r.matchScan(path, pos, &params)
	if node != nil {
		if n := node.findMethod(method); n != nil {
			return n, params
		}
	}
	return nil, nil
}

// matchScan 边扫描 path 边递归匹配
func (r *RadixNode) matchScan(path string, pos int, params *Params) *RadixNode {
	if pos >= len(path) {
		if r.hasMethod() {
			return r
		}
		if r.wildChild != nil {
			*params = append(*params, Param{Key: PathMatchVague, Value: ""})
			return r.wildChild
		}
		return nil
	}

	end := pos
	for end < len(path) && path[end] != '/' {
		end++
	}
	segment := path[pos:end]
	nextPos := end + 1
	if nextPos > len(path) {
		nextPos = len(path)
	}

	// 优先级 1：静态子节点（直接 slice 扫描，无 hash）
	if child := r.findStatic(toLowerFast(segment)); child != nil {
		if result := child.matchScan(path, nextPos, params); result != nil {
			return result
		}
	}

	// 优先级 2：参数子节点
	if r.paramChild != nil {
		mark := len(*params)
		for _, name := range r.paramChild.paramNames {
			*params = append(*params, Param{Key: name, Value: segment})
		}
		if result := r.paramChild.matchScan(path, nextPos, params); result != nil {
			return result
		}
		// 回溯：reslice 恢复，O(1)，无 map delete
		*params = (*params)[:mark]
	}

	// 优先级 3：通配符子节点
	if r.wildChild != nil {
		*params = append(*params, Param{Key: PathMatchVague, Value: path[pos:]})
		return r.wildChild
	}

	return nil
}

// toLowerFast 快速小写，大多数 URL 段已是小写时直接��回（零分配）
func toLowerFast(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			return Formatter(s)
		}
	}
	return s
}

// ==================== Static Router ====================

// Static 静态路由条目
type Static struct {
	methods []methodEntry
}

func (s *Static) findMethod(method string) *Node {
	for i := range s.methods {
		if s.methods[i].method == method {
			return s.methods[i].node
		}
	}
	return nil
}

// Router 路由器：静态路由走 O(1) hash map，动态路由走树匹配
type Router struct {
	static map[string]*Static
	radix  *RadixNode
}

func NewRouter() *Router {
	return &Router{
		static: make(map[string]*Static),
		radix:  NewRadixNode(""),
	}
}

// Static 静态路由查找
func (this *Router) Static(method string, path string) (route string, node *Node) {
	route = toLowerFast(path)
	static, ok := this.static[route]
	if !ok {
		return
	}
	node = static.findMethod(method)
	return
}

// Search 路由查��：静态优先，fallback 到树匹配
func (this *Router) Search(method string, paths ...string) (node *Node, params Params) {
	originalPath := Join(paths...)

	if _, node = this.Static(method, originalPath); node != nil {
		return node, nil
	}

	if this.radix != nil {
		n, p := this.radix.Match(originalPath, method)
		if n != nil {
			return n, p
		}
	}

	return nil, nil
}

// Register 注册路由
func (this *Router) Register(node *Node, method []string) (err error) {
	if len(method) == 0 {
		return fmt.Errorf("route register method empty:%s", node.Name())
	}
	route := node.Name()

	if !strings.Contains(route, PathMatchParam) && !strings.Contains(route, PathMatchVague) {
		static, ok := this.static[route]
		if !ok {
			static = &Static{}
			this.static[route] = static
		}
		for _, v := range method {
			v = strings.ToUpper(v)
			if static.findMethod(v) != nil {
				return fmt.Errorf("route exist:%s/%s", v, node.Name())
			}
			static.methods = append(static.methods, methodEntry{method: v, node: node})
		}
		return nil
	}

	return this.radix.Register(route, method, node)
}
