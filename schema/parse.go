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
		options:     opts,
		initialized: make(chan struct{}),
	}
	// Player exist schmema cache, return if exists
	if v, loaded := opts.Store.LoadOrStore(schemaCacheKey, schema); loaded {
		s := v.(*Schema)
		<-s.initialized
		return s, s.err
	} else {
		defer close(schema.initialized)
	}
	defer func() {
		if schema.err != nil {
			opts.Store.Delete(modelType)
		}
	}()

	modelValue := reflect.New(modelType)
	tableName := opts.TableName(modelType.Name())
	if tabler, ok := modelValue.Interface().(Tabler); ok {
		tableName = tabler.TableName()
	}
	if specialTableName != "" && specialTableName != tableName {
		tableName = specialTableName
	}

	schema.Name = modelType.Name()
	schema.ModelType = modelType
	schema.Table = tableName
	schema.FieldsByName = map[string]*Field{}
	schema.FieldsByDBName = map[string]*Field{}
	for i := 0; i < modelType.NumField(); i++ {
		fieldStruct := modelType.Field(i)
		if ast.IsExported(fieldStruct.Name) {
			field := schema.ParseField(fieldStruct)
			schema.Fields = append(schema.Fields, field.GetFields()...)
		}
	}

	for _, field := range schema.Fields {
		if field.DBName == "" {
			field.DBName = opts.ColumnName(schema.Table, field.Name)
		}

		if field.DBName != "" {
			// nonexistence or shortest path or first appear prioritized if has permission
			if _, ok := schema.FieldsByDBName[field.DBName]; !ok {
				schema.FieldsByDBName[field.DBName] = field
			}
		}

		if _, ok := schema.FieldsByName[field.Name]; !ok {
			schema.FieldsByName[field.Name] = field
		}

		field.setupValuerAndSetter()
	}

	return schema, schema.err
}

func getOrParse(dest interface{}, opts *Options) (*Schema, error) {
	modelType := reflect.ValueOf(dest).Type()
	for modelType.Kind() == reflect.Slice || modelType.Kind() == reflect.Array || modelType.Kind() == reflect.Ptr {
		modelType = modelType.Elem()
	}

	if modelType.Kind() != reflect.Struct {
		if modelType.PkgPath() == "" {
			return nil, fmt.Errorf("%w: %+v", ErrUnsupportedDataType, dest)
		}
		return nil, fmt.Errorf("%w: %s.%s", ErrUnsupportedDataType, modelType.PkgPath(), modelType.Name())
	}
	return ParseWithSpecialTableName(dest, "", opts)
}
