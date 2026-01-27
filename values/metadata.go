package values

import (
	"fmt"
	"strconv"
)

type Metadata map[string]string

func (meta Metadata) Set(k string, v any) {
	switch i := v.(type) {
	case string:
		meta[k] = i
	default:
		meta[k] = fmt.Sprintf("%v", v)
	}
}

func (meta Metadata) Get(keys ...string) (string, bool) {
	for _, k := range keys {
		if v, ok := meta[k]; ok {
			return v, true
		}
	}
	return "", false
}

func (meta Metadata) GetInt(k string) int {
	return int(meta.GetInt64(k))
}

func (meta Metadata) GetInt32(k string) int32 {
	return int32(meta.GetInt64(k))
}

func (meta Metadata) GetInt64(k string) int64 {
	s := meta[k]
	if s == "" {
		return 0
	}
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func (meta Metadata) GetUint(k string) uint {
	return uint(meta.GetUint64(k))
}

func (meta Metadata) GetUit32(k string) uint32 {
	return uint32(meta.GetUint64(k))
}

func (meta Metadata) GetUint64(k string) uint64 {
	s := meta[k]
	if s == "" {
		return 0
	}
	i, _ := strconv.ParseUint(s, 10, 64)
	return i
}

func (meta Metadata) GetFloat32(k string) float32 {
	return float32(meta.GetFloat64(k))
}

func (meta Metadata) GetFloat64(k string) (r float64) {
	s := meta[k]
	if s == "" {
		return 0
	}
	i, _ := strconv.ParseFloat(s, 64)
	return i
}
func (meta Metadata) GetString(k string) (r string) {
	return meta[k]
}

func (meta Metadata) Clone() Metadata {
	r := make(Metadata)
	for k, v := range meta {
		r[k] = v
	}
	return r
}
