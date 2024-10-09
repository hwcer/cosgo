package values

import (
	"encoding/json"
	"fmt"
	"github.com/hwcer/logger"
)

type Values map[string]any

func (vs Values) Key(k any) string {
	switch r := k.(type) {
	case string:
		return r
	default:
		return fmt.Sprint(r)
	}
}

func (vs Values) Has(k string) bool {
	_, ok := vs[k]
	return ok
}

func (vs Values) Get(k string) any {
	return vs[k]
}

// Set 保存数据，除了字符串和数字之外，一律转换成json字符串,Get时需要留意使用Unmarshal
func (vs Values) Set(k string, v any) any {
	switch v.(type) {
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, string:
		vs[k] = v
	default:
		vs.Marshal(k, v)
	}
	return vs[k]
}

func (vs Values) Clone() Values {
	r := make(Values, len(vs))
	for k, v := range vs {
		r[k] = v
	}
	return r
}

func (vs Values) GetInt(k string) int {
	return int(vs.GetInt64(k))
}

func (vs Values) GetInt32(k string) int32 {
	return int32(vs.GetInt64(k))
}

func (vs Values) GetInt64(k string) int64 {
	v, ok := vs[k]
	if !ok {
		return 0
	}
	return ParseInt64(v)
}

func (vs Values) GetFloat32(k string) float32 {
	return float32(vs.GetInt64(k))
}

func (vs Values) GetFloat64(k string) (r float64) {
	v, ok := vs[k]
	if !ok {
		return 0
	}
	return ParseFloat64(v)
}
func (vs Values) GetString(k string) (r string) {
	v, ok := vs[k]
	if !ok {
		return ""
	}
	return ParseString(v)
}

// Marshal 存入对象
func (vs Values) Marshal(k string, v any) {
	b, err := json.Marshal(v)
	if err == nil {
		vs[k] = string(b)
	} else {
		logger.Error("Values Marshal error,key:%v,val:%v,err:%v", k, v, err)
	}
}
func (vs Values) Unmarshal(k string, v any) error {
	s := vs.GetString(k)
	if s == "" {
		return nil
	}
	return json.Unmarshal([]byte(s), v)
}
