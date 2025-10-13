package registry

import (
	"reflect"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Join(paths ...string) (r string) {
	r = strings.Join(paths, "/")
	r = strings.Replace(r, "//", "/", -1)
	if !strings.HasPrefix(r, "/") {
		r = "/" + r
	}
	r = strings.TrimSuffix(r, "/")
	return
}

// Route 生成规范化的路由路径
func Route(paths ...string) string {
	var arr []string
	for _, v := range paths {
		if strings.HasPrefix(v, PathMatchParam) || strings.HasPrefix(v, PathMatchVague) {
			arr = append(arr, v)
		} else if v != "" && v != "/" {
			arr = append(arr, Formatter(v))
		}
	}
	if len(arr) == 0 {
		return "/"
	}
	return Formatter(Join(arr...))
}

//func Format(s string) string {
//	return strings.ToLower(s)
//}

func FuncName(i interface{}) (fname string) {
	fn := ValueOf(i)
	fname = runtime.FuncForPC(reflect.Indirect(fn).Pointer()).Name()
	if fname != "" {
		lastIndex := strings.LastIndex(fname, ".")
		if lastIndex >= 0 {
			fname = fname[lastIndex+1:]
		}
	}
	return strings.TrimSuffix(fname, "-fm")
}

func RouteName(name string) string {
	if strings.HasPrefix(name, PathMatchParam) {
		return PathMatchParam
	} else if strings.HasPrefix(name, PathMatchVague) {
		return PathMatchVague
	}
	return name
}

func ValueOf(i interface{}) reflect.Value {
	v, ok := i.(reflect.Value)
	if !ok {
		v = reflect.ValueOf(i)
	}
	return v
}

func IsExported(name string) bool {
	r, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(r)
}
