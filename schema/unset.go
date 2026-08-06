package schema

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Unset 按 mongo 风格的多级路径清除内存中的目标，是 $unset 落库后**重新读回**的等价态。
//
// 注意两侧动作并不相同：mongo 删的是文档里的字段，这里能做的只是把内存改成
// 「那个字段不存在时反序列化出来的样子」。$unset 本身不挑数据类型（数值同样适用，
// 文档中无任何按 BSON 类型的限制），下面每一行都按「删掉再读回来会是什么」推出：
//
//	目标            mongo 动作          读回 Go        本函数
//	结构体字段       删除字段            零值           置零值
//	子文档字段       删除该键            键消失         delete(m, key)   ← Go 的 map
//	数组元素        置 null，长度不变    []int64→0      置零值，长度不变
//	                                   []*T→nil
//
// 数组那行是 mongo 的显式设计:「replaces the matching element with null rather than
// removing... keeps consistent the array size and element positions」。
// []int64 里混进 null 读回来是 0、不报错，已用驱动 round-trip 实测过。
//
// 路径为空、字段不存在、已经钻到标量却还有剩余段 → 报错。
// 中途遇到 nil 指针 / nil map / 越界下标 → 视为「本来就没有」，直接返回 nil
// （与 $unset 对不存在字段是 no-op 一致）。
//
// 为什么不用 SetValue：它靠 embeddedSchema 下钻，只认结构体字段，遇到 map 直接放弃，
// 拿它删子键会报 "field not exist:a.b"。调用方（如 updater 的 dataset.Document）据此
// 认为内存已清，实则库里 $unset 成功、内存纹丝不动 —— 两边就此分叉，
// 常驻内存的模型会一直错到重启。
func (schema *Schema) Unset(obj any, path string) (err error) {
	defer func() {
		//下钻可能落到不可寻址的值上（如 map[K]Struct 的元素，Go 本身也不允许就地改），
		//reflect 会 panic；转成错误交给调用方，不要把整个请求带崩
		if e := recover(); e != nil {
			err = fmt.Errorf("schema unset panic,model:%s,path:%s,error:%v", schema.Name, path, e)
		}
	}()
	if path == "" {
		return fmt.Errorf("schema unset empty path,model:%s", schema.Name)
	}
	v := ValueOf(obj)
	rest := path
	//逐段下钻到**最后一段的父容器**，清除动作在父容器上做
	for {
		i := strings.IndexByte(rest, '.')
		if i < 0 {
			break
		}
		if rest[:i] == "" {
			return fmt.Errorf("schema unset path has empty segment,model:%s,path:%s", schema.Name, path)
		}
		if v, err = schema.unsetDescend(v, rest[:i], path); err != nil {
			return err
		}
		if !v.IsValid() {
			return nil //路径中断，本来就没有
		}
		rest = rest[i+1:]
	}
	if rest == "" {
		return fmt.Errorf("schema unset path has empty segment,model:%s,path:%s", schema.Name, path)
	}
	return schema.unsetClear(v, rest, path)
}

// unsetDescend 沿路径下钻一段，返回子值；子值不存在返回 zero Value（非错误）。
func (schema *Schema) unsetDescend(v reflect.Value, seg, path string) (reflect.Value, error) {
	if v = derefValue(v); !v.IsValid() {
		return v, nil
	}
	switch v.Kind() {
	case reflect.Struct:
		field, err := schema.fieldOfValue(v, seg, path)
		if err != nil {
			return reflect.Value{}, err
		}
		return field.Get(v), nil
	case reflect.Map:
		if v.IsNil() {
			return reflect.Value{}, nil
		}
		key, err := mapKeyOf(v.Type().Key(), seg)
		if err != nil {
			return reflect.Value{}, fmt.Errorf("schema unset bad map key,model:%s,segment:%s,path:%s,error:%v", schema.Name, seg, path, err)
		}
		return v.MapIndex(key), nil
	case reflect.Slice, reflect.Array:
		idx, err := strconv.Atoi(seg)
		if err != nil || idx < 0 || idx >= v.Len() {
			return reflect.Value{}, nil
		}
		return v.Index(idx), nil
	default:
		return reflect.Value{}, fmt.Errorf("schema unset path goes past a %s,model:%s,segment:%s,path:%s", v.Kind(), schema.Name, seg, path)
	}
}

// unsetClear 在父容器 v 上清除 seg 指向的目标。
func (schema *Schema) unsetClear(v reflect.Value, seg, path string) error {
	if v = derefValue(v); !v.IsValid() {
		return nil
	}
	switch v.Kind() {
	case reflect.Struct:
		field, err := schema.fieldOfValue(v, seg, path)
		if err != nil {
			return err
		}
		return field.Set(v, nil) //nil 走 fallbackSetter，置该字段类型的零值
	case reflect.Map:
		if v.IsNil() {
			return nil
		}
		key, err := mapKeyOf(v.Type().Key(), seg)
		if err != nil {
			return fmt.Errorf("schema unset bad map key,model:%s,segment:%s,path:%s,error:%v", schema.Name, seg, path, err)
		}
		v.SetMapIndex(key, reflect.Value{}) //零 Value = 删键
		return nil
	case reflect.Slice, reflect.Array:
		idx, err := strconv.Atoi(seg)
		if err != nil {
			return fmt.Errorf("schema unset bad index,model:%s,segment:%s,path:%s", schema.Name, seg, path)
		}
		if idx < 0 || idx >= v.Len() {
			return nil
		}
		e := v.Index(idx)
		if !e.CanSet() {
			return fmt.Errorf("schema unset unaddressable element,model:%s,path:%s", schema.Name, path)
		}
		e.Set(reflect.Zero(e.Type()))
		return nil
	default:
		return fmt.Errorf("schema unset path goes past a %s,model:%s,segment:%s,path:%s", v.Kind(), schema.Name, seg, path)
	}
}

// fieldOfValue 按结构体值的实际类型取字段。下钻穿过 map / slice 之后就没有现成的
// Schema 了，只能按类型查（parseType 命中全局缓存，不重复解析）。
func (schema *Schema) fieldOfValue(v reflect.Value, seg, path string) (*Field, error) {
	sub, err := parseType(v.Type(), schema.options, nil)
	if err != nil {
		return nil, err
	}
	field := sub.LookUpField(seg)
	if field == nil {
		return nil, fmt.Errorf("schema field not exist,model:%s,field:%s,path:%s", sub.Name, seg, path)
	}
	return field, nil
}

// derefValue 逐层解引用指针；遇到 nil 指针返回 zero Value（表示路径到此为止）。
func derefValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return reflect.Value{}
		}
		v = v.Elem()
	}
	return v
}

// mapKeyOf 把路径里的字符串段转成 map 的键类型。
// 只支持字符串与整型键 —— 其余（浮点、结构体键之类）在 mongo 的点号路径里也表达不出来。
func mapKeyOf(t reflect.Type, seg string) (reflect.Value, error) {
	k := reflect.New(t).Elem()
	switch t.Kind() {
	case reflect.String:
		k.SetString(seg)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(seg, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		if k.OverflowInt(n) {
			return reflect.Value{}, fmt.Errorf("key %s overflows %s", seg, t.Kind())
		}
		k.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(seg, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		if k.OverflowUint(n) {
			return reflect.Value{}, fmt.Errorf("key %s overflows %s", seg, t.Kind())
		}
		k.SetUint(n)
	default:
		return reflect.Value{}, fmt.Errorf("unsupported map key kind: %s", t.Kind())
	}
	return k, nil
}
