package internal

import (
	"fmt"
	"reflect"
	"strconv"
)

// 直接使用 strconv 进行转换，不使用缓存以避免内存使用不可控

// ToArray 将任意类型转换为 []interface{}
func ToArray(v interface{}) (r []interface{}) {
	vf := reflect.Indirect(reflect.ValueOf(v))
	if vf.Kind() != reflect.Array && vf.Kind() != reflect.Slice {
		return []interface{}{v}
	}
	for i := 0; i < vf.Len(); i++ {
		r = append(r, vf.Index(i).Interface())
	}
	return
}

// ToString 将任意类型转换为字符串
func ToString(value interface{}) string {
	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	default:
		return fmt.Sprintf("%v", value)
	}
}

// ToInt 将任意类型转换为 int64
func ToInt(i any) (r int64) {
	switch v := i.(type) {
	case int:
		r = int64(v)
	case uint:
		r = int64(v)
	case int8:
		r = int64(v)
	case uint8:
		r = int64(v)
	case int16:
		r = int64(v)
	case uint16:
		r = int64(v)
	case int32:
		r = int64(v)
	case uint32:
		r = int64(v)
	case int64:
		r = v
	case uint64:
		r = int64(v)
	case float32:
		r = int64(v)
	case float64:
		r = int64(v)
	case string:
		r, _ = strconv.ParseInt(v, 10, 64)
	}
	return
}

// ToInt32 将任意类型转换为 int32
func ToInt32(i any) int32 {
	return int32(ToInt(i))
}
