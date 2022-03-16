package utils

import (
	"fmt"
	"strconv"
)

type Dataset map[string]interface{}

func (m Dataset) Has(key string) bool {
	_, ok := m[key]
	return ok
}

func (m Dataset) Get(key string) interface{} {
	return m[key]
}

func (m Dataset) Set(key string, val interface{}) interface{} {
	m[key] = val
	return val
}

func (m Dataset) Add(key string, val int64) (r int64) {
	r = m.GetInt(key) + val
	m[key] = r
	return
}

func (m Dataset) Sub(key string, val int64) (r int64) {
	r = m.GetInt(key) - val
	m[key] = r
	return
}

func (m Dataset) GetInt(key string) int64 {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch v.(type) {
	case int:
		return int64(v.(int))
	case int32:
		return int64(v.(int32))
	case int64:
		return v.(int64)
	case float64:
		return int64(v.(float64))
	case string:
		temp, _ := strconv.ParseInt(v.(string), 10, 64)
		return temp
	default:
		return 0
	}
}
func (m Dataset) GetFloat(key string) (r float64) {
	v, ok := m[key]
	if !ok {
		return 0
	}
	switch v.(type) {
	case int:
		r = float64(v.(int))
	case int32:
		r = float64(v.(int32))
	case int64:
		r = float64(v.(int64))
	case float32:
		r = float64(v.(float32))
	case float64:
		r = v.(float64)
	case string:
		r, _ = strconv.ParseFloat(v.(string), 10)
	}
	return
}
func (m Dataset) GetString(key string) (r string) {
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
