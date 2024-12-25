package values

import (
	"fmt"
	"strconv"
)

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

func ParseFloat64(v any) (r float64) {
	if v == nil {
		return 0
	}
	switch v.(type) {
	case float32:
		r = float64(v.(float32))
	case float64:
		r = v.(float64)
	case string:
		r, _ = strconv.ParseFloat(v.(string), 10)
	default:
		return float64(ParseInt64(v))
	}
	return
}

func ParseString(v any) (r string) {
	if v == nil {
		return ""
	}
	switch v.(type) {
	case string:
		r = v.(string)
	default:
		r = fmt.Sprintf("%v", v)
	}
	return
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
