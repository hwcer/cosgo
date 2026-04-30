package internal

import (
	"reflect"
)

// Kind 获取接口或指针指向的底层类型
func Kind(dest any) reflect.Type {
	value := ValueOf(dest)
	if value.Kind() == reflect.Pointer && value.IsNil() {
		value = reflect.New(value.Type().Elem())
	}
	modelType := reflect.Indirect(value).Type()

	if modelType.Kind() == reflect.Interface {
		modelType = reflect.Indirect(reflect.ValueOf(dest)).Elem().Type()
	}

	for modelType.Kind() == reflect.Slice || modelType.Kind() == reflect.Array || modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
	}
	return modelType
}

// ValueOf 获取接口或值的 reflect.Value
func ValueOf(i any) reflect.Value {
	value, ok := i.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(i)
	}
	return value
}
