package values

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AKey interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

// Attach is not safe for concurrent use.
type Attach[K AKey] map[K]any

func (m Attach[K]) Has(k K) bool {
	_, ok := m[k]
	return ok
}

// Get 返回 key 对应的原始值。
// 复杂对象的实际存储类型不确定（可能是 bson.D/bson.A/[]byte 或已缓存的 Go 对象），
// 建议先使用 Unmarshal 反序列化为目标类型后再使用。
func (m Attach[K]) Get(k K) any {
	return m[k]
}

func (m Attach[K]) Set(k K, v any) any {
	m[k] = v
	return v
}

func (m Attach[K]) Range(f func(k K, v any) bool) {
	for k, v := range m {
		if !f(k, v) {
			return
		}
	}
}

func (m Attach[K]) Clone() Attach[K] {
	r := make(Attach[K], len(m))
	for k, v := range m {
		r[k] = v
	}
	return r
}

func (m Attach[K]) Merge(from Attach[K], replace bool) {
	for k, v := range from {
		if replace {
			m[k] = v
		} else if _, ok := m[k]; !ok {
			m[k] = v
		}
	}
}

func (m Attach[K]) GetInt(k K) int {
	return int(m.GetInt64(k))
}

func (m Attach[K]) GetInt32(k K) int32 {
	v, ok := m[k]
	if !ok {
		return 0
	}
	return ParseInt32(v)
}

func (m Attach[K]) GetInt64(k K) int64 {
	v, ok := m[k]
	if !ok {
		return 0
	}
	return ParseInt64(v)
}

func (m Attach[K]) GetFloat32(k K) float32 {
	v, ok := m[k]
	if !ok {
		return 0
	}
	return ParseFloat32(v)
}

func (m Attach[K]) GetFloat64(k K) float64 {
	v, ok := m[k]
	if !ok {
		return 0
	}
	return ParseFloat64(v)
}

func (m Attach[K]) GetString(k K) string {
	v, ok := m[k]
	if !ok {
		return ""
	}
	return ParseString(v)
}

// Unmarshal 将 key 对应的值反序列化到 i（必须是非空指针），返回深拷贝。
//
// 仅当值来自数据库或网络（bson.D/bson.A/[]byte）且需要反序列化为具体类型时才需要调用本方法。
// 如果数据不涉及序列化（如纯内存中通过 Set 写入的 Go 对象），直接使用 Get/GetInt/GetString 等方法即可。
//
// 推荐使用 Get + values.Unmarshal[V] + Set 替代本方法，泛型版本热路径零反射，性能更优：
//
//	result, err := values.Unmarshal[MyStruct](m.Get(k))
//	m.Set(k, result) // 缓存反序列化结果
//
// 本方法因 Go 暂不支持方法级别的类型参数，无法避免反射开销，待语言支持后改进。
func (m Attach[K]) Unmarshal(k K, i any) error {
	v := m[k]
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(i)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("values.Attach.Unmarshal: target must be a non-nil pointer, got %T", i)
	}
	switch raw := v.(type) {
	case bson.D, bson.A:
		// 冷路径：BSON 复合类型反序列化后缓存
		t, data, err := bson.MarshalValue(raw)
		if err != nil {
			return err
		}
		if err = bson.UnmarshalValue(t, data, i); err != nil {
			return err
		}
	case []byte:
		// 冷路径：旧版 JSON 字节反序列化后缓存
		if err := json.Unmarshal(raw, i); err != nil {
			return err
		}
	default:
		// 热路径：缓存命中，Go 对象深拷贝赋值
		src := reflect.ValueOf(v)
		if !src.Type().AssignableTo(rv.Type().Elem()) {
			return fmt.Errorf("values.Attach.Unmarshal: cannot assign %T to %s", v, rv.Type().Elem())
		}
		rv.Elem().Set(Clone(src, true))
		return nil
	}
	// 冷路径：缓存反序列化结果，后续调用走 default 热路径
	m[k] = Clone(rv.Elem(), true).Interface()
	return nil
}

func (m Attach[K]) MarshalJSON() ([]byte, error) {
	l := len(m)
	if l == 0 {
		return []byte("{}"), nil
	}
	b := bytes.NewBuffer([]byte("{"))
	var err error
	for k, v := range m {
		switch v := any(k).(type) {
		case string:
			_, err = fmt.Fprintf(b, `"%s":`, v)
		default:
			_, err = fmt.Fprintf(b, `"%d":`, v)
		}
		if err != nil {
			return nil, err
		}
		if raw, ok := v.([]byte); ok {
			_, err = b.Write(raw)
		} else if err = json.NewEncoder(b).Encode(v); err == nil && b.Len() > 0 && b.Bytes()[b.Len()-1] == '\n' {
			// bson.D/bson.A 实现了 json.Marshaler，json.NewEncoder 可正确序列化
			// json.Encoder.Encode 保证在末尾追加 \n（见 encoding/json/stream.go），需截断
			b.Truncate(b.Len() - 1)
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
