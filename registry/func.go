package registry

import (
	"reflect"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

func Join(paths ...string) string {

	r := strings.Join(paths, "/")
	// 处理重复的斜杠
	r = strings.Replace(r, "//", "/", -1)

	if !strings.HasPrefix(r, "/") {
		r = "/" + r
	}

	r = strings.TrimSuffix(r, "/")
	// 确保空路径返回根路径 "/"
	if r == "" {
		return "/"
	}
	return r
}

// Route 生成规范化的路由路径
func Route(paths ...string) string {
	// 预分配切片容量，减少扩容
	arr := make([]string, 0, len(paths))

	for _, v := range paths {
		if strings.Contains(v, PathMatchParam) || strings.Contains(v, PathMatchVague) {
			arr = append(arr, v)
		} else if v != "" && v != "/" {
			arr = append(arr, Formatter(v))
		}
	}

	if len(arr) == 0 {
		return "/"
	}

	return Join(arr...)
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

// Split 路径分割优化：减少内存分配
func Split(path string) []string {
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
	// 特殊处理根路径 "/"，返回 [""] 与 Register 方法保持一致
	if path == "/" {
		return []string{""}
	}
	return parts
}
