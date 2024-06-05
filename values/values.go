package values

import (
	"fmt"
	"strconv"
)

type Values map[string]interface{}

func (m Values) Has(key string) bool {
	_, ok := m[key]
	return ok
}

func (m Values) Get(key string) interface{} {
	return m[key]
}

func (m Values) Set(key string, val interface{}) interface{} {
	m[key] = val
	return val
}

//func (m Values) Add(key string, val int64) (r int64) {
//	r = m.GetInt64(key) + val
//	m[key] = r
//	return
//}
//
//func (m Values) Sub(key string, val int64) (r int64) {
//	r = m.GetInt64(key) - val
//	m[key] = r
//	return
//}

func (m Values) GetInt(key string) int {
	return int(m.GetInt64(key))
}

func (m Values) GetInt32(key string) int32 {
	return int32(m.GetInt64(key))
}

func (m Values) GetInt64(key string) int64 {
	v, ok := m[key]
	if !ok {
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
		return 0
	}
}

func (m Values) GetFloat32(key string) float32 {
	return float32(m.GetInt64(key))
}

func (m Values) GetFloat64(key string) (r float64) {
	v, ok := m[key]
	if !ok {
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
		return float64(m.GetInt64(key))
	}
	return
}
func (m Values) GetString(key string) (r string) {
	v, ok := m[key]
	if !ok {
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
