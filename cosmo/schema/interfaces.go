package schema

import (
	"gorm.io/gorm/clause"
)

type Tabler interface {
	TableName() string
}

type GormDataTypeInterface interface {
	GormDataType() string
}

//"BeforeCreate", "AfterCreate", "BeforeUpdate", "AfterUpdate", "BeforeDelete", "AfterDelete"

type BeforeCreate interface {
	BeforeCreate() []clause.Interface
}
