package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

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

func ValueOf(i interface{}) reflect.Value {
	value, ok := i.(reflect.Value)
	if !ok {
		value = reflect.ValueOf(i)
	}
	return value
}

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
func ToInt32(i any) int32 {
	return int32(ToInt(i))
}

func ParseTagSetting(str string, sep string) map[string]string {
	settings := map[string]string{}
	names := strings.Split(str, sep)

	for i := 0; i < len(names); i++ {
		j := i
		if len(names[j]) > 0 {
			for {
				if names[j][len(names[j])-1] == '\\' {
					i++
					names[j] = names[j][0:len(names[j])-1] + sep + names[i]
					names[i] = ""
				} else {
					break
				}
			}
		}

		values := strings.Split(names[j], ":")
		k := strings.TrimSpace(strings.ToUpper(values[0]))

		if len(values) >= 2 {
			settings[k] = strings.Join(values[1:], ":")
		} else if k != "" {
			settings[k] = k
		}
	}

	return settings
}
