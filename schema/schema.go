package schema

import (
	"fmt"
	"reflect"
	"strings"
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
	// indexes 缓存 ParseIndexes 的结果,避免每次调用都重新解析 tag。
	// Schema 构建完成后字段不变,索引结果也不变,安全缓存。
	indexes map[string]*Index
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
	return reflect.New(schema.ModelType)
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

// DBName 把多级路径逐段转换成**落库**字段名(bson)。
//
//	sch.DBName("SoulRelics")      → "soulrelics"
//	sch.DBName("SoulRelics.1")    → "soulrelics.1"     // 1 是 map 键,原样保留
//	sch.DBName("SoulRelics.1.Lv") → "soulrelics.1.lv"  // 按 map 值类型继续下钻
//
// 存在的意义:mongo 的 $set 对含 "." 的 key 原样下发、不查 schema,是整条写链路上
// 最后一次能纠正字段名的机会。大小写不符不会报错,只会在库里插一个野字段,
// 真正的字段纹丝不动 —— 静默丢数据。调用方(如 updater 的 Document)在把 key 交给
// ORM 之前过一遍这里,才谈得上「写进去的一定是库里那个字段」。
//
// 每一段按**当前容器类型**决定怎么处理:
//
//	结构体(含 *struct、匿名嵌入展开后) → 该段是字段名,查 schema 换名;查不到报错
//	map                              → 该段是 map 键,原样保留(键的大小写有业务含义)
//	slice / array                    → 该段是下标,原样保留
//	其它(标量等)                      → 已经到底却还有段,路径写错了,报错
func (schema *Schema) DBName(path string) (string, error) {
	return schema.GetName(path, "bson")
}

// JSName 把多级路径逐段转换成 **json** 字段名。
//
//	sch.JSName("SoulRelics.1.Lv") → "SoulRelics.1.lv"  // 各段取 json 标签,取不到用字段名
//
// 与 DBName 同一套逐段规则(见 DBName),差别只在每段取哪个名字。产出通常是发给客户端的
// key,同样不能原样透传 —— 客户端按 json 名认字段,大小写不符它就当这个字段不存在。
func (schema *Schema) JSName(path string) (string, error) {
	return schema.GetName(path, "json")
}

// GetName 把多级路径逐段转换成 tag 对应的字段名,是 DBName / JSName 的共同实现,
// 需要 bson / json 以外的标签(form 之类)时直接调它。逐段规则见 DBName。
//
// tag 取 "bson"(即 DBName)、"json"(即 JSName),或任意自定义标签名(走 Field.GetName,
// 取不到标签时回落成 Go 字段名)。注意 "bson" 取不到标签时回落成**小写**字段名,
// 与其它标签的回落规则不同 —— 那是 Field.DBName 的既定行为。
//
// 叫 GetName 而不是 Name,是因为 Schema.Name 已经是字段(模型名),方法不能同名;
// 与 Field.GetName、Schema.GetValue/SetValue 的命名也一致。
func (schema *Schema) GetName(path, tag string) (string, error) {
	i := strings.IndexByte(path, '.')
	if i < 0 {
		field := schema.LookUpField(path)
		if field == nil {
			return "", fmt.Errorf("schema field not exist,model:%s,field:%s", schema.Name, path)
		}
		return fieldName(field, tag), nil
	}
	field := schema.LookUpField(path[:i])
	if field == nil {
		return "", fmt.Errorf("schema field not exist,model:%s,field:%s,path:%s", schema.Name, path[:i], path)
	}

	//绝大多数 key 本来就是目标形态(业务照落库名写的 goods.10001),此时整条路径逐字不变。
	//所以 Builder 惰性启用:在出现第一处需要改名的段之前只记录「已匹配到哪」,
	//全程无差异就直接把入参原样返回 —— 零分配。
	//段也手动切,不走 strings.Split(那会为 []string 单独分配一次)。
	var b strings.Builder
	matched := 0 //b 未启用时,已产出的结果恒等于 path[:matched]
	if name := fieldName(field, tag); name == path[:i] {
		matched = i
	} else {
		b.Grow(len(path))
		b.WriteString(name)
	}

	//typ / sub 成对推进:sub 是 typ 对应的 Schema,由上一段的 Field.Embedded 直接给出
	//(解析期已备好),省掉按类型查全局缓存那一步。拿不到时为 nil,由 stepName 回落。
	typ, sub := field.IndirectFieldType, field.Embedded
	for pos := i + 1; ; {
		end := len(path)
		if j := strings.IndexByte(path[pos:], '.'); j >= 0 {
			end = pos + j
		}
		seg := path[pos:end]
		if seg == "" {
			return "", fmt.Errorf("schema path has empty segment,model:%s,path:%s", schema.Name, path)
		}
		name, next, nextSub, err := schema.stepName(typ, sub, seg, path, tag)
		if err != nil {
			return "", err
		}
		if b.Len() == 0 && name == seg {
			matched = end //仍与入参逐字相同,继续零分配
		} else {
			if b.Len() == 0 {
				b.Grow(len(path))
				b.WriteString(path[:matched]) //到此为止都没变过,整段前缀直接搬过来
			}
			b.WriteByte('.')
			b.WriteString(name)
		}
		typ, sub = next, nextSub
		if end == len(path) {
			break
		}
		pos = end + 1
	}
	if b.Len() == 0 {
		return path, nil
	}
	return b.String(), nil
}

// stepName 按当前容器类型解析路径中的一段,返回该段按 tag 转换后的名字,
// 以及下一段的容器类型与其对应的 Schema(拿不到时为 nil)。
//
// sub 是调用方已知的、与 typ 配对的 Schema —— 来自上一段 Field.Embedded,解析期就备好了。
// 有它就不必再按类型查一次全局缓存(实测 sync.Map 命中 12.75ns,直接取指针 0.2ns)。
// 二者必须成对传递:每一处产出 (typ, sub) 的地方都取自同一个 Field。
func (schema *Schema) stepName(typ reflect.Type, sub *Schema, seg, path, tag string) (string, reflect.Type, *Schema, error) {
	//map 值 / slice 元素的类型未经解引用,这里补上;Field.IndirectFieldType 则已在解析期去过指针
	for typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	switch typ.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		//map 键 / 下标原样保留。元素类型没有对应的 Field,给不出现成 Schema,
		//留给下一段按类型解析
		return seg, typ.Elem(), nil, nil
	case reflect.Struct:
		if sub == nil {
			//parseType 而非 GetOrParse:后者只收 any,得先 reflect.New 造个实例再把类型取回来,
			//白白多一次堆分配 —— 这里类型本来就在手上
			var err error
			if sub, err = parseType(typ, schema.options, nil); err != nil {
				return "", nil, nil, err
			}
		}
		field := sub.LookUpField(seg)
		if field == nil {
			return "", nil, nil, fmt.Errorf("schema field not exist,model:%s,field:%s,path:%s", sub.Name, seg, path)
		}
		return fieldName(field, tag), field.IndirectFieldType, field.Embedded, nil
	default:
		return "", nil, nil, fmt.Errorf("schema path goes past a %s,model:%s,segment:%s,path:%s", typ.Kind(), schema.Name, seg, path)
	}
}

// fieldName 取字段在 tag 下的名字。bson / json 走 Field 上带缓存的专用方法,
// 其余标签走通用的 GetName。
func fieldName(field *Field, tag string) string {
	switch tag {
	case "bson":
		return field.DBName()
	case "json":
		return field.JSName()
	default:
		return field.GetName(tag)
	}
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
	var sch *Schema
	for _, k := range keys {
		sk := ToString(k)
		sch = field.embeddedSchema()
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
	n := len(keys)
	vf = field.Get(vf)
	var sch *Schema
	for i := range n {
		sk := ToString(keys[i])
		sch = field.embeddedSchema()
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

// ParseField 解析单个结构体字段。公开 API,签名保持不变;环检测走内部的 parseField。
func (schema *Schema) ParseField(fieldStruct reflect.StructField) *Field {
	return schema.parseField(fieldStruct, nil)
}

// parseField 解析单个字段。chain 是当前 goroutine 的解析链,用于识别自引用 / 互引用。
func (schema *Schema) parseField(fieldStruct reflect.StructField, chain parsing) *Field {
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
	field.Index = field.StructField.Index
	if field.IndirectFieldType.Kind() != reflect.Struct {
		return field
	}

	//回指到解析链上的类型 = 自引用(Child *Node)或互引用(A.B→B.A)。
	//旧实现在这里无脑 GetOrParse,撞上缓存里那条属于自己的 in-progress 记录,
	//然后在 waitSchemaInit 里等一个只有自己才能 close 的 chan —— 一路等到
	//SchemaInitTimeout(默认 30s)才失败。
	if sub := chain.find(field.IndirectFieldType); sub != nil {
		if fieldStruct.Anonymous {
			//匿名嵌入要把子结构的字段**提升**到父结构上,而 processEmbeddedFields 是当场读
			//sub.Fields 的 —— 此刻它还没填完;何况自嵌入的提升本身就是无限的。
			//明确报错,不产出字段残缺的半成品。
			schema.setErr(fmt.Errorf("%w: %s <- %s", ErrEmbeddedCycle, schema.Name, sub.Name))
			return field
		}
		//具名字段:指针稳定,且没人会在解析期读它(提升只看匿名字段),
		//等到有人真去 LookUpField 时解析早已完成 —— 直接挂上去即可。
		field.Embedded = sub
		return field
	}

	//parseType 而非 GetOrParse:后者只收 any,得先 reflect.New 造个实例再把类型取回来
	sub, err := parseType(field.IndirectFieldType, schema.options, chain)
	if err != nil {
		schema.setErr(err)
		return field
	}
	field.Embedded = sub
	return field
}

// setErr 只记录**第一个**错误。
// 旧写法 `field.Embedded, schema.err = GetOrParse(...)` 是无条件赋值:第 2 个字段解析失败
// 置了 err,第 5 个字段成功又把它写回 nil,错误就丢了(processFields 每轮检查 schema.err,
// 恰好能逃过检查的组合确实存在)。
func (schema *Schema) setErr(err error) {
	if err != nil && schema.err == nil {
		schema.err = err
	}
}
