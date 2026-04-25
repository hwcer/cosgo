package registry

import (
	"reflect"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Join 将多段路径拼接并规范化为 "/a/b/c" 形式。
//
// Fast path: 单个参数、已 "/" 开头、无 "//"、无尾斜杠时直接返回,零分配。
// 典型 HTTP 请求 URL.Path 几乎都命中这一路径。
func Join(paths ...string) string {
	if len(paths) == 1 {
		p := paths[0]
		if p == "" {
			return "/"
		}
		if p[0] == '/' && !strings.Contains(p, "//") && (len(p) == 1 || p[len(p)-1] != '/') {
			return p
		}
	}
	r := strings.Join(paths, "/")
	r = strings.ReplaceAll(r, "//", "/")
	if !strings.HasPrefix(r, "/") {
		r = "/" + r
	}
	r = strings.TrimSuffix(r, "/")
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

// Split 按 '/' 分割路径,忽略空段。
// 根路径 "/" 返回 [""] 以与 Register 语义一致。
// 按 '/' 个数预估容量,单次分配,避免 append 扩容。
func Split(path string) []string {
	if path == "/" {
		return []string{""}
	}
	if path == "" {
		return nil
	}
	// '/' 个数上限即段数上限(实际可能少于,因首尾斜杠和空段)
	n := strings.Count(path, "/") + 1
	parts := make([]string, 0, n)
	var start int
	for i := 0; i < len(path); i++ {
		if path[i] == '/' {
			if i > start {
				parts = append(parts, path[start:i])
			}
			start = i + 1
		}
	}
	if start < len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}
