package registry

import (
	"reflect"
	"strings"
)


type RangeHandle func(name string, method interface{}) error
type FilterHandle func(reflect.Value, reflect.Value) bool

type Registry struct {
	*Prefix
	dict   map[string]*Route
	filter FilterHandle //用于判断struct中的方法是否合法接口
	Fuzzy  bool         //模糊匹配，不区分大小写
}

func New(prefix string, filter ...FilterHandle) *Registry {
	n := &Registry{
		dict:   make(map[string]*Route),
		Fuzzy:  true,
		Prefix: NewPrefix(prefix),
	}
	if len(filter) > 0 {
		n.filter = filter[0]
	}
	return n
}
func (this *Registry) Paths() (r []string) {
	prefix := this.Name()
	for k, _ := range this.dict {
		r = append(r, prefix+k)
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

func (this *Registry) Range(fn RangeHandle) (err error) {
	prefix := this.Name()
	for _, r := range this.dict {
		if err = r.Range(prefix, fn); err != nil {
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
		if pr, fn, ok = r.Match(path); ok {
			route = r
			return
		}
	}
	return
}
func (this *Registry) Format(path string) (new string) {
	new = Format(path)
	if this.Fuzzy {
		new = strings.ToLower(new)
	}
	return
}
