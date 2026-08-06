package schema

import (
	"fmt"
	"go/ast"
	"reflect"
	"time"

	"github.com/hwcer/logger"
)

// Parse 从目标结构体解析 Schema 信息
//
// 参数:
//
//	dest: 目标结构体实例或指针，用于解析其字段信息
//
// 返回:
//
//	*Schema: 解析生成的 Schema 实例，包含结构体的所有字段信息
//	error: 解析过程中发生的错误，如类型不支持等
//
// 示例:
//
//	type User struct {
//	    ID   int    `json:"id"`
//	    Name string `json:"name"`
//	}
//	schema, err := schema.Parse(&User{})
func Parse(dest any) (*Schema, error) {
	return ParseWithSpecialTableName(dest, "", config)
}

// GetOrParse 从目标结构体解析 Schema 信息，支持自定义选项
//
// 参数:
//
//	dest: 目标结构体实例或指针，用于解析其字段信息
//	opts: 解析选项，包含缓存、表名生成等配置
//
// 返回:
//
//	*Schema: 解析生成的 Schema 实例，包含结构体的所有字段信息
//	error: 解析过程中发生的错误，如类型不支持等
//
// 示例:
//
//	type User struct {
//	    ID   int    `json:"id"`
//	    Name string `json:"name"`
//	}
//	opts := &schema.Options{...}
//	schema, err := schema.GetOrParse(&User{}, opts)
func GetOrParse(dest any, opts *Options) (*Schema, error) {
	if opts == nil {
		opts = config
	}
	return ParseWithSpecialTableName(dest, "", opts)
}

// Warm 在应用启动阶段预先解析一批类型,以消除请求期首次访问时的 Schema 构建等待。
// 传入每个类型的零值指针(如 &User{}, &Order{}),使用默认 Options。
// 遇到任一解析错误立即返回该错误,已成功的 Schema 会保留在缓存中。
//
// 典型用法:
//
//	func main() {
//	    if err := schema.Warm(&User{}, &Order{}, &Product{}); err != nil {
//	        log.Fatal(err)
//	    }
//	    // ... 后续请求不会再触发同步构建等待
//	}
func Warm(dests ...any) error {
	return WarmWithOptions(config, dests...)
}

// WarmWithOptions 与 Warm 相同,但使用自定义 Options。
// 对每个 dest 调用 opts.Parse(dest),发生错误立即返回。
func WarmWithOptions(opts *Options, dests ...any) error {
	if opts == nil {
		opts = config
	}
	for _, d := range dests {
		if _, err := opts.Parse(d); err != nil {
			return fmt.Errorf("warm schema for %T: %w", d, err)
		}
	}
	return nil
}

// SchemaInitTimeout 等待其他 goroutine 完成 Schema 初始化的最大时长。
// 仅用于防御同 goroutine 自引用 struct 造成的死锁;正常并发构建等待是 μs 级,
// 基本不会触发此上限。超时后返回明确错误,避免静默卡住。
var SchemaInitTimeout = 30 * time.Second

// waitSchemaInit 等待 Schema 初始化完成并返回。
//
//   - 已完成(initDone 已 close): <-chan 立即返回 zero value,快路径无调度开销。
//   - 进行中: goroutine 阻塞在 chan 上,由 close 信号唤醒,O(μs) 级,不是 1ms 轮询。
//   - 自引用死锁/异常: SchemaInitTimeout 兜底返回错误。
//   - initDone == nil: 老版本遗留的 Schema(理论不出现),直接返回避免 panic。
func waitSchemaInit(s *Schema) (*Schema, error) {
	if s.initDone == nil {
		return s, s.err
	}
	// fast path: 已完成时 select default 分支不会命中,但 <-chan(已关闭) 立即就绪
	select {
	case <-s.initDone:
		return s, s.err
	default:
	}
	// slow path: 等待关闭 + 超时兜底
	timer := time.NewTimer(SchemaInitTimeout)
	defer timer.Stop()
	select {
	case <-s.initDone:
		return s, s.err
	case <-timer.C:
		return nil, fmt.Errorf("schema init timeout after %v (possible self-referential struct or concurrent init failure)", SchemaInitTimeout)
	}
}

// ParseWithSpecialTableName 从目标结构体解析 Schema 信息，支持自定义表名。
// 缓存命中时走零分配快路径（不设置 defer/recover），仅缓存未命中时进入完整解析。
func ParseWithSpecialTableName(dest any, specialTableName string, opts *Options) (*Schema, error) {
	if dest == nil {
		return nil, fmt.Errorf("%w: %+v", ErrUnsupportedDataType, dest)
	}

	modelType := Kind(dest)
	if err := validateStructType(modelType, dest); err != nil {
		return nil, err
	}

	schemaCacheKey := generateSchemaCacheKey(modelType, specialTableName)

	// 缓存命中快路径：无 defer、无 recover、零堆分配
	if v, ok := opts.Store.Load(schemaCacheKey); ok {
		return waitSchemaInit(v.(*Schema))
	}

	// 缓存未命中，进入完整解析（含 panic 保护）；顶层调用,解析链为空
	return parseSchemaSlow(modelType, schemaCacheKey, specialTableName, opts, nil)
}

// parsing 记录**当前 goroutine 这条解析链**上已经开工、尚未完成的 Schema。
//
// 为什么必须显式往下传:缓存里的 in-progress 记录,对**别的 goroutine** 是「schema 还没建好,
// 必须等」,对**本 goroutine** 才是「撞上环了」。二者语义相反,而 Go 没有 goroutine-local,
// 只能把链沿调用栈带下去。光靠 opts.Store 区分不出来 —— 这正是自引用结构会一路等到
// SchemaInitTimeout 的根因。
//
// 链深度一般 2~3 层,线性扫比 map 快,也省一次 map 分配。
type parsing []*Schema

// find 返回链上正在解析该类型的 Schema;不在链上返回 nil。
// ModelType 在 initializeSchemaBasicInfo 里就已写入,早于 processFields,故必定可比。
func (p parsing) find(t reflect.Type) *Schema {
	for _, s := range p {
		if s.ModelType == t {
			return s
		}
	}
	return nil
}

// parseType 与 GetOrParse 等价,但直接吃 reflect.Type。
//
// 调用方已经握有类型时(按 map 值 / slice 元素类型逐段下钻就是这种),用它可以省掉
// 「reflect.New 造一个实例 → Kind() 再把类型取回来」这一次纯浪费的堆分配。
// 缓存键与 generateSchemaCacheKey(modelType, "") 一致,与 Parse 共用同一份缓存。
//
// chain 为本次解析链,顶层调用传 nil。
func parseType(modelType reflect.Type, opts *Options, chain parsing) (*Schema, error) {
	if opts == nil {
		opts = config
	}
	for modelType.Kind() == reflect.Pointer {
		modelType = modelType.Elem()
	}
	if modelType.Kind() != reflect.Struct {
		return nil, fmt.Errorf("%w: %v", ErrUnsupportedDataType, modelType)
	}
	if v, ok := opts.Store.Load(modelType); ok {
		return waitSchemaInit(v.(*Schema))
	}
	return parseSchemaSlow(modelType, modelType, "", opts, chain)
}

// parseSchemaSlow 缓存未命中时的完整解析路径，独立函数以隔离 defer/recover 开销。
// chain 是调用方那条解析链，本层会把自己追加进去再往下传，用于环检测。
func parseSchemaSlow(modelType reflect.Type, cacheKey any, specialTableName string, opts *Options, chain parsing) (res *Schema, err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("schema parse panic: %v", e)
			logger.Alert("schema parse panic: %v", e)
		}
	}()

	schema := &Schema{
		options:  opts,
		initDone: make(chan struct{}),
	}

	if actual, loaded := opts.Store.LoadOrStore(cacheKey, schema); loaded {
		return waitSchemaInit(actual.(*Schema))
	}

	defer close(schema.initDone)
	defer func() {
		if schema.err != nil {
			opts.Store.Delete(cacheKey)
		}
	}()

	if err := initializeSchemaBasicInfo(schema, modelType, specialTableName); err != nil {
		return nil, err
	}
	//本层入链后再解析字段：字段里若回指到链上任一类型，即为自引用/互引用，由 parseField 处理。
	//必须在 initializeSchemaBasicInfo 之后，ModelType 那时才写好、可供比对。
	//
	//三下标切片强制每次都复制：否则 len<cap 时兄弟字段的递归会共享底层数组、
	//在同一下标上互相覆写。当前是串行的、覆写也正确，但那是「靠时序才对」——
	//解析是冷路径，一次小分配换掉这份推演负担。
	if err := processFields(schema, modelType, append(chain[:len(chain):len(chain)], schema)); err != nil {
		return nil, err
	}
	processEmbeddedFields(schema)
	if err := buildFieldMappings(schema); err != nil {
		return nil, err
	}
	return schema, schema.err
}

// validateStructType 验证目标类型是否为结构体
func validateStructType(modelType reflect.Type, dest any) error {
	if modelType.Kind() != reflect.Struct {
		if modelType.PkgPath() == "" {
			return NewUnsupportedDataTypeError(dest)
		}
		return fmt.Errorf("%w: %s.%s", ErrUnsupportedDataType, modelType.PkgPath(), modelType.Name())
	}
	// 允许临时结构体（anonymous struct）通过验证
	// 临时结构体的 PkgPath() 为空字符串
	return nil
}

// schemaCacheKey 在有自定义表名时作为 sync.Map 键,避免 fmt.Sprintf 分配字符串。
// reflect.Type 实现了可比较接口,与 name 一起构成可比较的 struct,能直接作 map key。
type schemaCacheKey struct {
	t    reflect.Type
	name string
}

// generateSchemaCacheKey 生成 Schema 缓存键。
// 无自定义表名时直接用 reflect.Type(canonical);有表名则用 struct 组合,零字符串分配。
func generateSchemaCacheKey(modelType reflect.Type, specialTableName string) any {
	if specialTableName != "" {
		return schemaCacheKey{t: modelType, name: specialTableName}
	}
	return modelType
}

// initializeSchemaBasicInfo 初始化 Schema 基本信息
func initializeSchemaBasicInfo(schema *Schema, modelType reflect.Type, specialTableName string) error {
	// 确定表名
	tableName := determineTableName(modelType, specialTableName, schema.options)

	// 设置基本信息
	schema.Name = modelType.Name()
	schema.ModelType = modelType
	schema.Table = tableName

	// 预分配 map 容量,避免频繁扩容
	fieldCount := modelType.NumField()
	schema.Fields = make(map[string]*Field, fieldCount)

	return nil
}

// determineTableName 确定表名
func determineTableName(modelType reflect.Type, specialTableName string, opts *Options) string {
	// 优先使用用户指定的表名
	if specialTableName != "" {
		return specialTableName
	}

	// 尝试从 Tabler 接口获取表名
	modelValue := reflect.New(modelType)
	if tabler, ok := modelValue.Interface().(Tabler); ok {
		if tableName := tabler.TableName(); tableName != "" {
			return tableName
		}
	}

	// 使用默认表名生成规则
	return opts.TableName(modelType.Name())
}

// processFields 处理结构体字段。chain 为含本层在内的解析链，用于环检测。
func processFields(schema *Schema, modelType reflect.Type, chain parsing) error {
	for i := range modelType.NumField() {
		fieldStruct := modelType.Field(i)
		if ast.IsExported(fieldStruct.Name) {
			field := schema.parseField(fieldStruct, chain)

			if field.StructField.Anonymous {
				schema.Embedded = append(schema.Embedded, field)
			} else {
				schema.Fields[field.Name] = field
			}
		}
		if schema.err != nil {
			return schema.err
		}
	}
	return nil
}

// processEmbeddedFields 处理嵌入字段
func processEmbeddedFields(schema *Schema) {
	for _, v := range schema.Embedded {
		for _, field := range v.GetEmbeddedFields() {
			if _, ok := schema.Fields[field.Name]; !ok {
				schema.Fields[field.Name] = field
			}
		}
	}
}

// buildFieldMappings 一次遍历 schema.Fields,构建:
//   - unifiedFields: 按 Go 名 > db 标签 > json 标签 的优先级合并的统一索引
//   - dbFields: 带 db 标签的字段列表(供 Schema.Range 高效迭代)
//   - db 名重复检测通过本地临时 map 完成,结束即弃,不占 Schema 字段空间
func buildFieldMappings(schema *Schema) error {
	n := len(schema.Fields)
	// unifiedFields 容量最多是 n*3(每字段最多 3 个 alias),保守预分配
	unified := make(map[string]*Field, n*2)
	// 先以 Go 名占位(最高优先级),避免后续 db/json alias 覆盖
	for k, v := range schema.Fields {
		unified[k] = v
	}
	// dbSeen 仅用于 dup 检测,构建完即 GC;dbFields 保留给 Schema.Range
	dbSeen := make(map[string]*Field, n)
	schema.dbFields = make([]*Field, 0, n)

	for _, field := range schema.Fields {
		if db := field.DBName(); db != "" {
			if f, exists := dbSeen[db]; exists {
				// 同一字段在 Fields 里可能因多次索引出现(理论上 Fields key 去重,这里保险)
				if f != field {
					return NewDuplicateDBNameError(schema.Name, f.Name, field.Name)
				}
			} else {
				dbSeen[db] = field
				schema.dbFields = append(schema.dbFields, field)
				if _, taken := unified[db]; !taken {
					unified[db] = field
				}
			}
		}
		if js := field.JSName(); js != "" {
			if _, taken := unified[js]; !taken {
				unified[js] = field
			}
		}
	}
	schema.unifiedFields = unified
	// 预热每个字段的 getValueFn/setValueFn,避免运行期首次访问的 lazy compile 抖动,
	// 并消除并发首次访问双 compile 的浪费;嵌入提升字段因已重置指针,此处重新编译校正 Index。
	warmupFieldFns(schema)
	return nil
}

// warmupFieldFns 预先编译所有可访问字段的 getValueFn/setValueFn。
// 包括直接字段和经匿名嵌入提升到本 Schema 的字段 —— 后者已在 GetEmbeddedFields
// 中重置过 fn 指针,此处按调整后的 Index 重新编译,保证访问正确。
func warmupFieldFns(schema *Schema) {
	for _, f := range schema.Fields {
		if f.getValueFn == nil {
			f.compileGetValueFn()
		}
		if f.setValueFn == nil {
			f.compileSetValueFn()
		}
	}
}
