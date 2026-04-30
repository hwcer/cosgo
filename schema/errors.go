package schema

import (
	"errors"
	"fmt"
)

var (
	ErrUnsupportedDataType = errors.New("unsupported data type")
	ErrDuplicateDBName     = errors.New("duplicate database field name")
)

func NewUnsupportedDataTypeError(dest any) error {
	return fmt.Errorf("%w: %+v", ErrUnsupportedDataType, dest)
}

func NewDuplicateDBNameError(structName, fieldName1, fieldName2 string) error {
	return fmt.Errorf("%w: struct(%s) DBName repeat: %s,%s", ErrDuplicateDBName, structName, fieldName1, fieldName2)
}
