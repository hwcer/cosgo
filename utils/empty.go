package utils

import "reflect"

const (
	IntZero     = int(0)
	Int8Zero    = int8(0)
	Int16Zero   = int16(0)
	Int32Zero   = int32(0)
	Int64Zero   = int64(0)
	UintZero    = uint(0)
	Uint8Zero   = uint8(0)
	Uint16Zero  = uint16(0)
	Uint32Zero  = uint32(0)
	Uint64Zero  = uint64(0)
	Float32Zero = float32(0)
	Float64Zero = float64(0)
	StringZero  = ""
)

func IsEmpty(data interface{}) bool {
	if data == nil {
		return false
	}
	switch v := data.(type) {
	case bool:
		return false
	case *bool:
		return v == nil
	case int:
		return v == IntZero
	case *int:
		return v == nil || *v == IntZero
	case int8:
		return v == Int8Zero
	case *int8:
		return v == nil || *v == Int8Zero
	case int16:
		return v == Int16Zero
	case *int16:
		return v == nil || *v == Int16Zero
	case int32:
		return v == Int32Zero
	case *int32:
		return v == nil || *v == Int32Zero
	case int64:
		return v == Int64Zero
	case *int64:
		return v == nil || *v == Int64Zero
	case uint:
		return v == UintZero
	case *uint:
		return v == nil || *v == UintZero
	case uint8:
		return v == Uint8Zero
	case *uint8:
		return v == nil || *v == Uint8Zero
	case uint16:
		return v == Uint16Zero
	case *uint16:
		return v == nil || *v == Uint16Zero
	case uint32:
		return v == Uint32Zero
	case *uint32:
		return v == nil || *v == Uint32Zero
	case uint64:
		return v == Uint64Zero
	case *uint64:
		return v == nil || *v == Uint64Zero
	case float32:
		return v == Float32Zero
	case *float32:
		return v == nil || *v == Float32Zero
	case float64:
		return v == Float64Zero
	case *float64:
		return v == nil || *v == Float64Zero
	case string:
		return v == StringZero
	case *string:
		return v == nil || *v == StringZero
	default:
		return IsEmptyReflect(data)
	}
}

// IsEmptyReflect 通过反射判断复杂类型是否为空,一般确定参数不为基础类型时才直接使用，否则应该使用 IsEmpty
func IsEmptyReflect(object interface{}) bool {
	// get nil case out of the way
	if object == nil {
		return true
	}
	objValue := reflect.ValueOf(object)
	switch objValue.Kind() {
	// collection types are empty when they have no element
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice:
		return objValue.Len() == 0
		// pointers are empty if nil or if the value they point to is empty
	case reflect.Ptr:
		if objValue.IsNil() {
			return true
		}
		deref := objValue.Elem().Interface()
		return IsEmptyReflect(deref)
		// for all other types, compare against the zero value
	default:
		zero := reflect.Zero(objValue.Type())
		return reflect.DeepEqual(object, zero.Interface())
	}
}
