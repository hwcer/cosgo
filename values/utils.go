package values

import (
	"fmt"
	"strconv"
)

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
