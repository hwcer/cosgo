package schema

import (
	"reflect"

	"github.com/hwcer/cosgo/schema/internal"
)

// Kind 获取接口或指针指向的底层类型
func Kind(dest interface{}) reflect.Type {
	return internal.Kind(dest)
}

// ValueOf 获取接口或值的 reflect.Value
func ValueOf(i interface{}) reflect.Value {
	return internal.ValueOf(i)
}

// ToArray 将任意类型转换为 []interface{}
func ToArray(v interface{}) (r []interface{}) {
	return internal.ToArray(v)
}

// ToString 将任意类型转换为字符串
func ToString(value interface{}) string {
	return internal.ToString(value)
}

// ToInt 将任意类型转换为 int64
func ToInt(i any) (r int64) {
	return internal.ToInt(i)
}

// ToInt32 将任意类型转换为 int32
func ToInt32(i any) int32 {
	return internal.ToInt32(i)
}

// ParseTagSetting 解析结构体标签设置
func ParseTagSetting(str string, sep string) map[string]string {
	return internal.ParseTagSetting(str, sep)
}
