package internal

import (
	"reflect"
)

// ReflectType 是 reflect.Type 的别名
type ReflectType = reflect.Type

// ReflectValue 是 reflect.Value 的别名
type ReflectValue = reflect.Value

// Kind 获取接口或指针指向的底层类型
func Kind(dest interface{}) reflect.Type {
	value := ValueOf(dest)
	if value.Kind() == reflect.Ptr && value.IsNil() {
		value = reflect.New(value.Type().Elem())
	}
	modelType := reflect.Indirect(value).Type()

	if modelType.Kind() == reflect.Interface {
		modelType = reflect.Indirect(reflect.ValueOf(dest)).Elem().Type()
	}

	for modelType.Kind() == reflect.Slice || modelType.Kind() == reflect.Array || modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}
	return modelType
}

// ValueOf 获取接口或值的 reflect.Value
func ValueOf(i interface{}) reflect.Value {
	value, ok := i.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(i)
	}
	return value
}
