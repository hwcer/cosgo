package schema

import (
	"errors"
	"fmt"
)

// 错误定义
var (
	// ErrUnsupportedDataType 不支持的数据类型
	ErrUnsupportedDataType = errors.New("unsupported data type")
	// ErrDuplicateFieldName 字段名重复
	ErrDuplicateFieldName = errors.New("duplicate field name")
	// ErrDuplicateDBName 数据库字段名重复
	ErrDuplicateDBName = errors.New("duplicate database field name")
	// ErrInvalidTag 无效的标签格式
	ErrInvalidTag = errors.New("invalid tag format")
	// ErrFieldNotFound 字段不存在
	ErrFieldNotFound = errors.New("field not found")
	// ErrNotObject 不是对象类型
	ErrNotObject = errors.New("not an object")
)

// 错误构造函数

// NewUnsupportedDataTypeError 创建不支持的数据类型错误
func NewUnsupportedDataTypeError(dest interface{}) error {
	return fmt.Errorf("%w: %+v", ErrUnsupportedDataType, dest)
}

// NewDuplicateFieldNameError 创建字段名重复错误
func NewDuplicateFieldNameError(structName, fieldName string) error {
	return fmt.Errorf("%w: struct(%s) field name repeat: %s", ErrDuplicateFieldName, structName, fieldName)
}

// NewDuplicateDBNameError 创建数据库字段名重复错误
func NewDuplicateDBNameError(structName, fieldName1, fieldName2 string) error {
	return fmt.Errorf("%w: struct(%s) DBName repeat: %s,%s", ErrDuplicateDBName, structName, fieldName1, fieldName2)
}

// NewInvalidTagError 创建无效的标签格式错误
func NewInvalidTagError(fieldName, tag string) error {
	return fmt.Errorf("%w: field(%s) tag format error: %s", ErrInvalidTag, fieldName, tag)
}

// NewFieldNotFoundError 创建字段不存在错误
func NewFieldNotFoundError(path string) error {
	return fmt.Errorf("%w: %s", ErrFieldNotFound, path)
}

// NewNotObjectError 创建不是对象类型错误
func NewNotObjectError(path string) error {
	return fmt.Errorf("%w: %s", ErrNotObject, path)
}
