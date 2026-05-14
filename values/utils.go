package values

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Unmarshal 将 raw (bson.D/bson.A/[]byte 或已解码的 Go 对象) 反序列化为 V。
// 若 raw 已经是 V 类型则直接返回（零反射）。
func Unmarshal[V any](raw any) (r V, err error) {
	if raw == nil {
		return
	}
	if v, ok := raw.(V); ok {
		return v, nil
	}
	switch v := raw.(type) {
	case bson.D, bson.A:
		t, data, e := bson.MarshalValue(v)
		if e != nil {
			return r, e
		}
		err = bson.UnmarshalValue(t, data, &r)
	case []byte:
		err = json.Unmarshal(v, &r)
	default:
		err = fmt.Errorf("values.Unmarshal: cannot unmarshal %T into %T", raw, r)
	}
	return
}

// Clone 复制 reflect.Value。deep 为 true 时递归深拷贝，否则浅拷贝。
func Clone(src reflect.Value, deep bool) reflect.Value {
	if !deep {
		dst := reflect.New(src.Type()).Elem()
		dst.Set(src)
		return dst
	}
	return deepClone(src)
}

func deepClone(src reflect.Value) reflect.Value {
	switch src.Kind() {
	case reflect.Pointer:
		if src.IsNil() {
			return reflect.Zero(src.Type())
		}
		dst := reflect.New(src.Type().Elem())
		dst.Elem().Set(deepClone(src.Elem()))
		return dst
	case reflect.Slice:
		if src.IsNil() {
			return reflect.Zero(src.Type())
		}
		dst := reflect.MakeSlice(src.Type(), src.Len(), src.Len())
		for i := range src.Len() {
			dst.Index(i).Set(deepClone(src.Index(i)))
		}
		return dst
	case reflect.Map:
		if src.IsNil() {
			return reflect.Zero(src.Type())
		}
		dst := reflect.MakeMapWithSize(src.Type(), src.Len())
		iter := src.MapRange()
		for iter.Next() {
			dst.SetMapIndex(iter.Key(), deepClone(iter.Value()))
		}
		return dst
	case reflect.Struct:
		dst := reflect.New(src.Type()).Elem()
		for i := range src.NumField() {
			if dst.Field(i).CanSet() {
				dst.Field(i).Set(deepClone(src.Field(i)))
			}
		}
		return dst
	case reflect.Interface:
		if src.IsNil() {
			return reflect.Zero(src.Type())
		}
		return deepClone(src.Elem())
	default:
		return src
	}
}

func ParseInt32(v any) int32 {
	return int32(ParseInt64(v))
}

func ParseInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch d := v.(type) {
	case int:
		return int64(d)
	case uint:
		return int64(d)
	case int8:
		return int64(d)
	case uint8:
		return int64(d)
	case int16:
		return int64(d)
	case uint16:
		return int64(d)
	case int32:
		return int64(d)
	case uint32:
		return int64(d)
	case int64:
		return int64(d)
	case uint64:
		return int64(d)
	case float32:
		return int64(d)
	case float64:
		return int64(d)
	case string:
		temp, _ := strconv.ParseInt(d, 10, 64)
		return temp
	default:
		temp, _ := strconv.ParseInt(fmt.Sprintf("%v", d), 10, 64)
		return temp
	}
}

func ParseFloat32(v any) (r float32) {
	return float32(ParseFloat64(v))
}
func ParseFloat64(v any) float64 {
	if v == nil {
		return 0
	}
	switch d := v.(type) {
	case float32:
		return float64(d)
	case float64:
		return d
	case string:
		r, _ := strconv.ParseFloat(d, 64)
		return r
	default:
		return float64(ParseInt64(v))
	}
}
func ParseString(v any) string {
	if v == nil {
		return ""
	}
	switch d := v.(type) {
	case string:
		return d
	default:
		return fmt.Sprintf("%v", d)
	}
}

func Sprintf(format any, args ...any) (text string) {
	switch v := format.(type) {
	case string:
		text = v
	case error:
		text = v.Error()
	default:
		text = fmt.Sprintf("%v", format)
	}
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}
	return
}

func IsBasicType(v interface{}) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr, float32, float64, string, bool:
		return true
	default:
		return false
	}
}
