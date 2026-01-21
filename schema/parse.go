package schema

import (
	"fmt"
	"go/ast"
	"reflect"
	"sync/atomic"
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
func Parse(dest interface{}) (*Schema, error) {
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
func GetOrParse(dest interface{}, opts *Options) (*Schema, error) {
	return ParseWithSpecialTableName(dest, "", opts)
}

// waitSchemaInit 等待 Schema 初始化完成并返回
func waitSchemaInit(s *Schema) (*Schema, error) {
	// 等待初始化完成，使用原子操作确保内存可见性
	for atomic.LoadUint32(&s.initState) != 2 {
		// 使用短暂休眠替代纯粹的忙等，减少 CPU 占用
		// 当多次循环时，逐渐增加休眠时间，避免频繁调度
		time.Sleep(1 * time.Nanosecond)
	}
	return s, s.err
}

// ParseWithSpecialTableName 从目标结构体解析 Schema 信息，支持自定义表名
//
// 参数:
//
//	dest: 目标结构体实例或指针，用于解析其字段信息
//	specialTableName: 自定义表名，优先级高于结构体的 Tabler 接口
//	opts: 解析选项，包含缓存、表名生成等配置
//
// 返回:
//
//	*Schema: 解析生成的 Schema 实例，包含结构体的所有字段信息
//	error: 解析过程中发生的错误，如类型不支持、字段名重复等
//
// 示例:
//
//	type User struct {
//	    ID   int    `json:"id" db:"user_id"`
//	    Name string `json:"name" db:"user_name"`
//	}
//	opts := &schema.Options{...}
//	schema, err := schema.ParseWithSpecialTableName(&User{}, "custom_users", opts)
func ParseWithSpecialTableName(dest interface{}, specialTableName string, opts *Options) (*Schema, error) {
	if dest == nil {
		return nil, fmt.Errorf("%w: %+v", ErrUnsupportedDataType, dest)
	}
	defer func() {
		if e := recover(); e != nil {
			logger.Debug(e)
		}
	}()

	// 验证目标类型是否为结构体
	modelType := Kind(dest)
	if err := validateStructType(modelType, dest); err != nil {
		return nil, err
	}

	// 生成缓存键
	schemaCacheKey := generateSchemaCacheKey(modelType, specialTableName)

	// 创建并初始化 Schema 实例
	schema := &Schema{
		options:   opts,
		initState: 0,
	}

	// 检查缓存并处理并发
	if cachedSchema, loaded := checkSchemaCache(opts, schemaCacheKey, schema); loaded {
		return waitSchemaInit(cachedSchema)
	}

	// 缓存不存在，设置初始化状态
	atomic.StoreUint32(&schema.initState, 1)

	// 初始化完成后更新状态
	defer func() {
		atomic.StoreUint32(&schema.initState, 2)
	}()

	// 初始化出错时从缓存删除
	defer func() {
		if schema.err != nil {
			opts.Store.Delete(schemaCacheKey)
		}
	}()

	// 初始化 Schema 基本信息
	if err := initializeSchemaBasicInfo(schema, modelType, specialTableName); err != nil {
		return nil, err
	}

	// 处理字段
	if err := processFields(schema, modelType); err != nil {
		return nil, err
	}

	// 处理嵌入字段
	processEmbeddedFields(schema)

	// 构建字段映射
	if err := buildFieldMappings(schema); err != nil {
		return nil, err
	}

	return schema, schema.err
}

// validateStructType 验证目标类型是否为结构体
func validateStructType(modelType reflect.Type, dest interface{}) error {
	if modelType.Kind() != reflect.Struct {
		if modelType.PkgPath() == "" {
			return NewUnsupportedDataTypeError(dest)
		}
		return fmt.Errorf("%w: %s.%s", ErrUnsupportedDataType, modelType.PkgPath(), modelType.Name())
	}
	return nil
}

// generateSchemaCacheKey 生成 Schema 缓存键
func generateSchemaCacheKey(modelType reflect.Type, specialTableName string) interface{} {
	if specialTableName != "" {
		return fmt.Sprintf("%p-%s", modelType, specialTableName)
	}
	return modelType
}

// checkSchemaCache 检查 Schema 缓存并处理并发
func checkSchemaCache(opts *Options, schemaCacheKey interface{}, schema *Schema) (*Schema, bool) {
	if v, loaded := opts.Store.LoadOrStore(schemaCacheKey, schema); loaded {
		return v.(*Schema), true
	}
	return nil, false
}

// initializeSchemaBasicInfo 初始化 Schema 基本信息
func initializeSchemaBasicInfo(schema *Schema, modelType reflect.Type, specialTableName string) error {
	// 确定表名
	tableName := determineTableName(modelType, specialTableName, schema.options)

	// 设置基本信息
	schema.Name = modelType.Name()
	schema.ModelType = modelType
	schema.Table = tableName

	// 预分配 map 容量，避免频繁扩容
	fieldCount := modelType.NumField()
	schema.Fields = make(map[string]*Field, fieldCount)
	schema.fieldsJson = make(map[string]*Field, fieldCount)
	schema.fieldsDatabase = make(map[string]*Field, fieldCount)

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

// processFields 处理结构体字段
func processFields(schema *Schema, modelType reflect.Type) error {
	for i := 0; i < modelType.NumField(); i++ {
		fieldStruct := modelType.Field(i)
		if ast.IsExported(fieldStruct.Name) {
			field := schema.ParseField(fieldStruct)

			if field.StructField.Anonymous {
				schema.Embedded = append(schema.Embedded, field)
			} else {
				schema.fieldsPrivate = append(schema.fieldsPrivate, field)
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

// buildFieldMappings 构建字段映射
func buildFieldMappings(schema *Schema) error {
	for _, field := range schema.Fields {
		// 构建数据库字段映射
		if db := field.DBName(); db != "" {
			if f, ok := schema.fieldsDatabase[db]; !ok {
				schema.fieldsDatabase[db] = field
			} else {
				return NewDuplicateDBNameError(schema.Name, f.Name, field.Name)
			}
		}
		// 构建 JSON 字段映射
		if jsName := field.JSName(); jsName != "" {
			schema.fieldsJson[jsName] = field
		}
	}
	return nil
}
