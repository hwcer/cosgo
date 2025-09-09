package values

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson/primitive"
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
	vs[k] = v
	return v
}
func (vs Values) Range(f func(k string, v any) bool) {
	for k, v := range vs {
		if !f(k, v) {
			return
		}
	}
}
func (vs Values) Clone() Values {
	r := make(Values, len(vs))
	for k, v := range vs {
		r[k] = v
	}
	return r
}

func (vs Values) Merge(from Values, replace bool) {
	for k, v := range from {
		if replace {
			vs[k] = v
		} else if _, ok := vs[k]; !ok {
			vs[k] = v
		}
	}
}

func (vs Values) GetInt(k string) int {
	return int(vs.GetInt64(k))
}

func (vs Values) GetInt32(k string) int32 {
	v, ok := vs[k]
	if !ok {
		return 0
	}
	return ParseInt32(v)
}

func (vs Values) GetInt64(k string) int64 {
	v, ok := vs[k]
	if !ok {
		return 0
	}
	return ParseInt64(v)
}

func (vs Values) GetFloat32(k string) float32 {
	v, ok := vs[k]
	if !ok {
		return 0
	}
	return ParseFloat32(v)
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

func (vs Values) Marshal(k string, v any) (r any, err error) {
	if IsBasicType(v) {
		vs[k] = v
		return v, nil
	}
	var b []byte
	if b, err = json.Marshal(v); err == nil {
		r = Bytes(b)
		vs[k] = r
	}
	return
}
func (vs Values) Unmarshal(k string, i any) error {
	v := vs[k]
	if v == nil {
		return nil
	}
	switch s := v.(type) {
	case primitive.Binary:
		return json.Unmarshal(s.Data, i)
	case []byte:
		return json.Unmarshal(s, i)
	case Bytes:
		return json.Unmarshal(s, i)
	default:
		return errors.New("invalid type")
	}
}
func (vs Values) MarshalJSON() ([]byte, error) {
	l := len(vs)
	if l == 0 {
		return []byte("{}"), nil
	}
	b := bytes.NewBuffer([]byte("{"))
	je := json.NewEncoder(b)
	var err error
	for k, v := range vs {
		b.WriteString(fmt.Sprintf(`"%s":`, k))
		switch i := v.(type) {
		case []byte:
			_, err = b.Write(i)
		case primitive.Binary:
			_, err = b.Write(i.Data)
		case Bytes:
			_, err = b.Write(i)
		default:
			err = je.Encode(i)
		}
		if err != nil {
			return nil, err
		}
		l--
		if l > 0 {
			b.WriteString(",")
		}
	}
	b.WriteString("}")

	return b.Bytes(), nil
}
