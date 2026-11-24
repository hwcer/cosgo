package values

import (
	"bytes"
	"encoding/json"
	"fmt"
	"maps"
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
	maps.Copy(r, m)
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

// Unmarshal 把 key 对应的值反序列化到 i（必须是非空指针），返回深拷贝：读出的值可以随意修改。
// 只读场景用 Borrow，不拷贝。
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
func (m Attach[K]) Unmarshal(k K, i any) error { return m.Load(k, i, true) }

// Borrow 与 Unmarshal 同形，但不深拷贝：把 key 对应的值按原形交给 i 所指的变量，供只读场景使用。
//
// 读出的值与缓存共享内部引用——切片 / map / 指针就是缓存里那一份，结构体本身是浅拷贝。
// 调用方不得修改读出的值及其内部引用，否则会污染内存里的缓存（请求回滚后内存与库不一致）；
// 要改就用 Unmarshal 取深拷贝。标量读取两者没有区别。
func (m Attach[K]) Borrow(k K, i any) error { return m.Load(k, i, false) }

// Load 把 key 对应的值读进 i（必须是非空指针）；没有这个值时目标保持不变。
//
// deep 决定读出的值与缓存是否独立：true 深拷贝，读出的值可以修改（Unmarshal）；
// false 不拷贝，切片 / map / 指针与缓存共享，只能读（Borrow）。
//
// 缓存里的值有两种形态：
//
//   - bson.D / bson.A、[]byte：刚从库里读出，还没解码（冷路径）。解码进 i，并把解码结果记入缓存，
//     之后的读取走热路径。
//   - Go 对象（热路径）：直接赋给 i。Set 原样保存调用方给的对象（业务侧通常持有 *T），
//     冷路径记入的却是读入目标所指的类型，两者差一层指针时自动对齐。
func (m Attach[K]) Load(k K, i any, deep bool) error {
	v := m[k]
	if v == nil {
		return nil //没有这个值
	}
	rv := reflect.ValueOf(i)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return fmt.Errorf("values.Attach: target must be a non-nil pointer, got %T", i)
	}
	dst := rv.Elem()
	switch raw := v.(type) {
	case bson.D, bson.A:
		// 冷路径
		if err := m.decodeBSON(raw, i); err != nil {
			return err
		}
	case []byte:
		// 冷路径：旧版 JSON 字节
		if err := json.Unmarshal(raw, i); err != nil {
			return err
		}
	default:
		// 热路径
		return m.assign(dst, reflect.ValueOf(v), deep)
	}
	// 把解码结果记入缓存，之后的读取走热路径。
	// deep：克隆一份与调用方目标独立的副本；否则直接存，缓存与这次读出的值共享——Borrow 的语义本就如此。
	if deep {
		m[k] = Clone(dst, true).Interface()
	} else {
		m[k] = dst.Interface()
	}
	return nil
}

// 以下是 Load 用到的内部步骤，只服务于 Attach，不进包命名空间。

// decodeBSON 把缓存里的 BSON 复合值反序列化进 i。
func (m Attach[K]) decodeBSON(raw any, i any) error {
	t, data, err := bson.MarshalValue(raw)
	if err != nil {
		return err
	}
	return bson.UnmarshalValue(t, data, i)
}

// assign 热路径：把缓存里的 Go 对象 src 赋给目标 dst。deep 为 true 深拷贝，否则按原形浅赋值。
// 类型对不上时兼容差一层指针（Set 原样保存调用方给的对象，业务侧通常持有 *T；冷路径记入的却是读入目标
// 所指的类型）；各 case 的条件都是「这样对齐之后能否赋值」，全都不成立才报错。
func (m Attach[K]) assign(dst, src reflect.Value, deep bool) error {
	switch {
	case src.Type().AssignableTo(dst.Type()): //同型
	case src.Kind() == reflect.Pointer && src.IsNil():
		return nil //存的是空指针 = 没有值，目标保持不变
	case src.Kind() == reflect.Pointer && dst.Kind() != reflect.Pointer && src.Elem().Type().AssignableTo(dst.Type()):
		src = src.Elem() //存指针、读值
	case src.Kind() != reflect.Pointer && dst.Kind() == reflect.Pointer && reflect.PointerTo(src.Type()).AssignableTo(dst.Type()):
		p := reflect.New(src.Type()) //存值、读指针：给出一份浅拷贝的地址
		p.Elem().Set(src)
		src = p
	default:
		return fmt.Errorf("values.Attach: cannot assign %s to %s", src.Type(), dst.Type())
	}
	if deep {
		dst.Set(Clone(src, true))
	} else {
		dst.Set(src)
	}
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
		//key必须经json.Marshal转义,含引号/反斜杠/控制字符的key会产生非法JSON
		switch key := any(k).(type) {
		case string:
			kb, e := json.Marshal(key)
			if e != nil {
				return nil, e
			}
			_, _ = b.Write(kb)
			_, _ = b.WriteString(":")
		default:
			_, err = fmt.Fprintf(b, `"%d":`, key)
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
