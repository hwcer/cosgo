package registry

import "reflect"

type Options struct {
	Format func(string) string                     //格式化路径
	Filter func(reflect.Value, reflect.Value) bool //用于判断struct中的方法是否合法接口
}
