package registry

import (
	"reflect"
	"strings"
)

type RegistryRangeHandle func(path string, method interface{}) error
type RegistryFilterHandle func(reflect.Value, reflect.Value) bool

type Registry struct {
	dict   map[string]*Route
	filter RegistryFilterHandle //用于判断struct中的方法是否合法接口
	Fuzzy  bool                 //模糊匹配，不区分大小写
}

func New(filter ...RegistryFilterHandle) *Registry {
	n := &Registry{
		dict:  make(map[string]*Route),
		Fuzzy: true,
	}
	if len(filter) > 0 {
		n.filter = filter[0]
	}
	return n
}
func (this *Registry) Nodes() (r []string) {
	for k, _ := range this.dict {
		r = append(r, k)
	}
	return
}
func (this *Registry) Route(name string) *Route {
	route := NewRoute(this, name)

	if r, ok := this.dict[route.name]; ok {
		return r
	}
	this.dict[route.name] = route
	return route
}

func (this *Registry) Range(fn RegistryRangeHandle) (err error) {
	for _, r := range this.dict {
		if err = this.rangeRoute(r, fn); err != nil {
			return
		}
	}
	return
}

func (this *Registry) Match(path string) (route *Route, pr, fn reflect.Value, ok bool) {
	path = this.Format(path)
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	for _, r := range this.dict {
		if pr, fn, ok = r.match(path); ok {
			route = r
			return
		}
	}
	return
}
func (this *Registry) Format(path string) (new string) {
	new = strings.Trim(path, "/")
	if this.Fuzzy {
		new = strings.ToLower(new)
	}
	return
}

func (this *Registry) rangeRoute(r *Route, fn RegistryRangeHandle) (err error) {
	for _, n := range r.nodes {
		if err = this.rangeNode(r, n, fn); err != nil {
			return
		}
	}
	for k, m := range r.method {
		name := this.pathName(r.name, k)
		if err = fn(name, m); err != nil {
			return
		}
	}
	return nil
}

func (this *Registry) rangeNode(route *Route, node *Node, fn RegistryRangeHandle) (err error) {
	for k, m := range node.method {
		name := this.pathName(route.name, node.name, k)
		if err = fn(name, m); err != nil {
			return
		}
	}
	return nil
}

func (this *Registry) pathName(prefix string, path ...string) string {
	if prefix == "/" {
		prefix = ""
	}
	return strings.Join(append([]string{prefix}, path...), "/")
}
