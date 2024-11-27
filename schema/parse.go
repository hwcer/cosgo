package schema

import (
	"fmt"
	"go/ast"
	"reflect"
)

// Parse get data type from dialector
func Parse(dest interface{}) (*Schema, error) {
	return ParseWithSpecialTableName(dest, "", config)
}
func GetOrParse(dest interface{}, opts *Options) (*Schema, error) {
	return ParseWithSpecialTableName(dest, "", opts)
}

// ParseWithSpecialTableName get data type from dialector with extra schema table
func ParseWithSpecialTableName(dest interface{}, specialTableName string, opts *Options) (*Schema, error) {
	if dest == nil {
		return nil, fmt.Errorf("%w: %+v", ErrUnsupportedDataType, dest)
	}
	modelType := Kind(dest)
	if modelType.Kind() != reflect.Struct {
		if modelType.PkgPath() == "" {
			return nil, fmt.Errorf("%w: %+v", ErrUnsupportedDataType, dest)
		}
		return nil, fmt.Errorf("%w: %s.%s", ErrUnsupportedDataType, modelType.PkgPath(), modelType.Name())
	}

	// Cache the schema for performance,
	// Use the modelType or modelType + schemaTable (if it present) as cache key.
	var schemaCacheKey interface{}
	if specialTableName != "" {
		schemaCacheKey = fmt.Sprintf("%p-%s", modelType, specialTableName)
	} else {
		schemaCacheKey = modelType
	}

	schema := &Schema{
		init:    make(chan struct{}),
		options: opts,
	}
	// Player exist schmema cache, return if exists
	if v, loaded := opts.Store.LoadOrStore(schemaCacheKey, schema); loaded {
		s := v.(*Schema)
		if s.init != nil {
			<-s.init
		}
		return s, s.err
	} else {
		defer schema.initialized()
	}
	defer func() {
		if schema.err != nil {
			opts.Store.Delete(modelType)
		}
	}()

	modelValue := reflect.New(modelType)
	var tableName string
	if tabler, ok := modelValue.Interface().(Tabler); ok {
		tableName = tabler.TableName()
	} else {
		tableName = opts.TableName(modelType.Name())
	}
	if specialTableName != "" && specialTableName != tableName {
		tableName = specialTableName
	}

	schema.Name = modelType.Name()
	schema.ModelType = modelType
	schema.Table = tableName
	schema.Fields = map[string]*Field{}
	schema.fieldsDatabase = map[string]*Field{}

	//var embeddedField []*Field

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
			return nil, schema.err
		}
	}
	//所有Anonymous字段

	for _, v := range schema.Embedded {
		for _, field := range v.GetEmbeddedFields() {
			if _, ok := schema.Fields[field.Name]; !ok {
				schema.Fields[field.Name] = field
			}
		}
	}

	for _, field := range schema.Fields {
		if field.DBName != "" {
			if f, ok := schema.fieldsDatabase[field.DBName]; !ok {
				schema.fieldsDatabase[field.DBName] = field
			} else {
				return nil, fmt.Errorf("struct(%v) DBName repeat: %v,%v", schema.Name, f.Name, field.Name)
			}
		}
	}

	return schema, schema.err
}
