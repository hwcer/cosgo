package schema

type Tabler interface {
	TableName() string
}

type GormDataTypeInterface interface {
	GormDataType() string
}
