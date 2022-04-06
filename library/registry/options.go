package registry

import (
	"path"
	"reflect"
	"strings"
)

func NewOptions() *Options {
	return &Options{route: map[string]*Service{}}
}

type Options struct {
	Format func(string) string                     //格式化路径
	Filter func(reflect.Value, reflect.Value) bool //用于判断struct中的方法是否合法接口
	route  map[string]*Service
}

//Clean 将所有path 格式化成 /a/b 模式
func (this *Options) Clean(paths ...string) (r string) {
	p := path.Join(paths...)
	if this.Format != nil {
		r = this.Format(p)
	} else {
		r = strings.ToLower(p)
	}
	r = strings.Trim(r, "/")
	if r != "" {
		r = strings.Join([]string{"/", r}, "")
	}
	return
}

func (this *Options) addRoutePath(r *Service, s ...string) {
	p := []string{r.name}
	p = append(p, s...)
	k := path.Join(p...)
	this.route[k] = r
}
