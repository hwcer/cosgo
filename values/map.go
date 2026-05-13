package values

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MapKey interface {
	~string | ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64
}

type Map[K MapKey] map[K]any

func (m Map[K]) Has(k K) bool {
	_, ok := m[k]
	return ok
}

func (m Map[K]) Get(k K) any {
	return m[k]
}

func (m Map[K]) Set(k K, v any) any {
	m[k] = v
	return v
}

func (m Map[K]) Range(f func(k K, v any) bool) {
	for k, v := range m {
		if !f(k, v) {
			return
		}
	}
}

func (m Map[K]) Clone() Map[K] {
	r := make(Map[K], len(m))
	for k, v := range m {
		r[k] = v
	}
	return r
}

func (m Map[K]) Merge(from Map[K], replace bool) {
	for k, v := range from {
		if replace {
			m[k] = v
		} else if _, ok := m[k]; !ok {
			m[k] = v
		}
	}
}

func (m Map[K]) GetInt32(k K) int32 {
	v, ok := m[k]
	if !ok {
		return 0
	}
	return ParseInt32(v)
}

func (m Map[K]) GetInt64(k K) int64 {
	v, ok := m[k]
	if !ok {
		return 0
	}
	return ParseInt64(v)
}

func (m Map[K]) GetFloat32(k K) float32 {
	v, ok := m[k]
	if !ok {
		return 0
	}
	return ParseFloat32(v)
}

func (m Map[K]) GetFloat64(k K) float64 {
	v, ok := m[k]
	if !ok {
		return 0
	}
	return ParseFloat64(v)
}

func (m Map[K]) GetString(k K) string {
	v, ok := m[k]
	if !ok {
		return ""
	}
	return ParseString(v)
}

// reflectCopy 复制 src 返回独立副本，slice/map 深拷贝底层数据，其余走值语义
func reflectCopy(src reflect.Value) reflect.Value {
	switch src.Kind() {
	case reflect.Slice:
		dst := reflect.MakeSlice(src.Type(), src.Len(), src.Len())
		reflect.Copy(dst, src)
		return dst
	case reflect.Map:
		dst := reflect.MakeMapWithSize(src.Type(), src.Len())
		iter := src.MapRange()
		for iter.Next() {
			dst.SetMapIndex(iter.Key(), iter.Value())
		}
		return dst
	default:
		return src
	}
}

// Unmarshal 将 key 对应的值反序列化到 i（必须是非空指针）。
func (m Map[K]) Unmarshal(k K, i any) error {
	v := m[k]
	if v == nil {
		return nil
	}
	// 1. 内存中的 Go 对象：类型匹配时复制值（服务器运行时主要路径）
	rv := reflect.ValueOf(i)
	if rv.Kind() == reflect.Pointer && !rv.IsNil() {
		src := reflect.ValueOf(v)
		if src.Type().AssignableTo(rv.Type().Elem()) {
			rv.Elem().Set(reflectCopy(src))
			return nil
		}
	}
	// 2. DB 读出的 BSON 复合类型：值级别反序列化，缓存独立副本替换原始 BSON 值
	switch v.(type) {
	case bson.D, bson.A:
		t, data, err := bson.MarshalValue(v)
		if err != nil {
			return err
		}
		if err = bson.UnmarshalValue(t, data, i); err != nil {
			return err
		}
		if rv.Kind() == reflect.Pointer && !rv.IsNil() {
			m[k] = reflectCopy(rv.Elem()).Interface()
		}
		return nil
	}
	// 3. 旧版兼容：历史数据中 Marshal 存储的 JSON 字节
	if s, ok := v.([]byte); ok {
		return json.Unmarshal(s, i)
	}
	return fmt.Errorf("values.Map.Unmarshal: unsupported type %T", v)
}

func (m Map[K]) MarshalJSON() ([]byte, error) {
	l := len(m)
	if l == 0 {
		return []byte("{}"), nil
	}
	b := bytes.NewBuffer([]byte("{"))
	je := json.NewEncoder(b)
	je.SetEscapeHTML(false)
	var err error
	for k, v := range m {
		switch i := v.(type) {
		case []byte:
			fmt.Fprintf(b, `"%v":`, k)
			_, err = b.Write(i)
		case bson.D, bson.A:
			var d []byte
			if d, err = bson.MarshalExtJSON(bson.D{{Key: fmt.Sprintf("%v", k), Value: v}}, false, false); err == nil {
				_, err = b.Write(d[1 : len(d)-1])
			}
		default:
			fmt.Fprintf(b, `"%v":`, k)
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
