package schema

import (
	"fmt"
	"reflect"
)

type Schema struct {
	err       error
	options   *Options
	Name      string
	Table     string
	Embedded  []*Field //匿名嵌入字段的 Field 列表
	ModelType reflect.Type
	// Fields 按 Go 字段名(含嵌入提升)索引的全量字段表,公开访问。
	Fields map[string]*Field
	// unifiedFields 同时包含 Go 名、db 标签、json 标签作为 key 的统一索引。
	// LookUpField 走它实现单次 map 查询。优先级: Go 名 > db > json。
	unifiedFields map[string]*Field
	// dbFields 保留带 db 标签字段的列表,供 Schema.Range 高效迭代(slice 迭代比 map 快)。
	// 顺序与 Fields 迭代一致,但只包含 DBName() != "" 的字段。
	dbFields []*Field
	// initDone 在 Schema 创建时分配,构建者在 Parse 返回前 close()。
	// 并发等待者通过 <-initDone 阻塞到构建完成,O(μs) 唤醒;
	// 构建完成后 close 的 chan 对后续读者零开销(立即返回 zero value)。
	initDone chan struct{}
}

func (schema *Schema) String() string {
	if schema.ModelType.Name() == "" {
		return fmt.Sprintf("%s(%s)", schema.Name, schema.Table)
	}
	return fmt.Sprintf("%s.%s", schema.ModelType.PkgPath(), schema.ModelType.Name())
}

func (schema *Schema) New() reflect.Value {
	results := reflect.New(schema.ModelType)
	return results
}

// Make Make a Slice
func (schema *Schema) Make() reflect.Value {
	slice := reflect.MakeSlice(reflect.SliceOf(reflect.PointerTo(schema.ModelType)), 0, 20)
	results := reflect.New(slice.Type())
	results.Elem().Set(slice)
	return results
}

// Range 按注册顺序遍历所有带 db 标签的字段(用于 ORM 持久化场景)。
// slice 迭代比 map 更快,且顺序稳定。
func (schema *Schema) Range(cb func(*Field) bool) {
	for _, field := range schema.dbFields {
		if !cb(field) {
			return
		}
	}
}

// LookUpField 按名字查找字段,支持 Go 字段名、db 标签名、json 标签名任一(优先级依此顺序)。
// 通过构建期合并的 unifiedFields 单次 map 查询完成。
func (schema *Schema) LookUpField(name string) *Field {
	return schema.unifiedFields[name]
}
func (schema *Schema) JSName(k string) (r string) {
	if field := schema.LookUpField(k); field != nil {
		return field.JSName()
	}
	return
}
// GetValue 按路径取值。单 key 的 common case(len(keys)==0)走快路径,不分配合并 slice。
func (schema *Schema) GetValue(obj any, key string, keys ...any) (r any) {
	vf := ValueOf(obj)
	// 第一段 key 单独处理,避免 append([]any{key}, keys...) 分配
	field := schema.LookUpField(key)
	if field == nil {
		return nil
	}
	vf = field.Get(vf)
	if !vf.IsValid() {
		return nil
	}
	if len(keys) == 0 {
		return vf.Interface()
	}
	// 多段路径,走通用逻辑
	var sch *Schema
	for i := 0; i < len(keys); i++ {
		sk := ToString(keys[i])
		// 对于指针类型的嵌入字段,获取其 Schema
		if field.FieldType.Kind() == reflect.Pointer && field.IndirectFieldType.Kind() == reflect.Struct {
			tempValue := reflect.New(field.IndirectFieldType)
			tempSchema, err := GetOrParse(tempValue.Interface(), field.Schema.options)
			if err != nil {
				return nil
			}
			sch = tempSchema
		} else {
			sch = field.Embedded
		}
		if sch == nil {
			return nil
		}
		field = sch.LookUpField(sk)
		if field == nil {
			return nil
		}
		vf = field.Get(vf)
		if !vf.IsValid() {
			return nil
		}
	}
	return vf.Interface()
}

// SetValue 按路径赋值。单 key 的 common case 走快路径,不分配合并 slice。
func (schema *Schema) SetValue(obj any, val any, key string, keys ...any) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	vf := ValueOf(obj)
	field := schema.LookUpField(key)
	if field == nil {
		return fmt.Errorf("field not exist:%v", key)
	}
	if len(keys) == 0 {
		return field.Set(vf, val)
	}
	// 多段路径: 中间段 Get 向下钻,末段 Set
	n := len(keys)
	vf = field.Get(vf)
	var sch *Schema
	for i := 0; i < n; i++ {
		sk := ToString(keys[i])
		sch = field.Embedded
		if sch == nil {
			return fmt.Errorf("field not object at %v", key)
		}
		field = sch.LookUpField(sk)
		if field == nil {
			return fmt.Errorf("field not exist:%v", sk)
		}
		if i < n-1 {
			vf = field.Get(vf)
		}
	}
	return field.Set(vf, val)
}

func (schema *Schema) ParseField(fieldStruct reflect.StructField) *Field {
	//var err error

	field := &Field{
		Name:              fieldStruct.Name,
		FieldType:         fieldStruct.Type,
		IndirectFieldType: fieldStruct.Type,
		StructField:       fieldStruct,
		Schema:            schema,
	}

	for field.IndirectFieldType.Kind() == reflect.Pointer {
		field.IndirectFieldType = field.IndirectFieldType.Elem()
	}
	fieldValue := reflect.New(field.IndirectFieldType)
	field.Index = field.StructField.Index
	kind := reflect.Indirect(fieldValue).Kind()
	switch kind {
	case reflect.Struct:
		field.Embedded, schema.err = GetOrParse(fieldValue.Interface(), schema.options)
	case reflect.Map, reflect.Slice, reflect.Array:
		//初始化子结构
	case reflect.Invalid, reflect.Uintptr, reflect.Chan, reflect.Func, reflect.Interface, reflect.UnsafePointer, reflect.Complex64, reflect.Complex128:
		//schema.err = fmt.Errorf("invalid embedded struct for %s's field %s, should be struct, but got %v", field.Schema.Name, field.Name, field.FieldType)
	default:

	}
	return field
}
