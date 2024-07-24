package values

import (
	"fmt"
	"strconv"
)

type Values map[string]any

func (m Values) rk(k any) string {
	switch r := k.(type) {
	case string:
		return r
	case int:
		return strconv.Itoa(r)
	case int32:
		return strconv.Itoa(int(r))
	case int64:
		return strconv.FormatInt(r, 10)
	default:
		return fmt.Sprint(r)
	}
}

func (m Values) Has(k any) bool {
	rk := m.rk(k)
	_, ok := m[rk]
	return ok
}

func (m Values) Get(k any) any {
	rk := m.rk(k)
	return m[rk]
}

func (m Values) Set(k any, v any) {
	rk := m.rk(k)
	m[rk] = v
}

func (m Values) GetInt(k any) int {
	return int(m.GetInt64(k))
}

func (m Values) GetInt32(k any) int32 {
	return int32(m.GetInt64(k))
}

func (m Values) GetInt64(k any) int64 {
	rk := m.rk(k)
	v, ok := m[rk]
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

func (m Values) GetFloat32(k any) float32 {
	return float32(m.GetInt64(k))
}

func (m Values) GetFloat64(k any) (r float64) {
	rk := m.rk(k)
	v, ok := m[rk]
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
		return float64(m.GetInt64(k))
	}
	return
}
func (m Values) GetString(k any) (r string) {
	rk := m.rk(k)
	v, ok := m[rk]
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
