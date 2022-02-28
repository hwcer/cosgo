package registry

import (
	"reflect"
	"runtime"
	"strings"
	"unicode"
	"unicode/utf8"
)

func FuncName(i interface{}) (fname string) {
	fn := ValueOf(i)
	fname = runtime.FuncForPC(reflect.Indirect(fn).Pointer()).Name()
	if fname != "" {
		lastIndex := strings.LastIndex(fname, ".")
		if lastIndex >= 0 {
			fname = fname[lastIndex+1:]
		}
	}
	return
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

//Format 格式化路径为 /x/y模式
func Format(name string) string {
	s := strings.Trim(name, "/")
	if s == "" {
		return s
	} else {
		return "/" + s
	}
}
