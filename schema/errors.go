package schema

import (
	"errors"
	"fmt"
)

var (
	ErrUnsupportedDataType = errors.New("unsupported data type")
	ErrDuplicateDBName     = errors.New("duplicate database field name")
	// ErrEmbeddedCycle 匿名嵌入构成环。具名字段的自引用/互引用是合法的(照常解析),
	// 但匿名嵌入要把子结构的字段提升到父结构上,成环即无限提升,只能报错。
	ErrEmbeddedCycle = errors.New("embedded field cycle")
)

func NewUnsupportedDataTypeError(dest any) error {
	return fmt.Errorf("%w: %+v", ErrUnsupportedDataType, dest)
}

func NewDuplicateDBNameError(structName, fieldName1, fieldName2 string) error {
	return fmt.Errorf("%w: struct(%s) DBName repeat: %s,%s", ErrDuplicateDBName, structName, fieldName1, fieldName2)
}
