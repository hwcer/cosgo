package schema

type Tabler interface {
	TableName() string
}

type GetIndexes interface {
	GetIndexes() []*IndexField
}
