package schema

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// https://www.mongodb.com/zh-cn/docs/manual/core/index-partial/#std-label-index-type-partial
// 部分索引 使用一条查询语句
const (
	IndexTag      = "index"
	IndexName     = "NAME"
	IndexSort     = "SORT"
	IndexUnique   = "UNIQUE"
	IndexSparse   = "SPARSE"
	IndexPartial  = "PARTIAL" //部分索引 index:"PARTIAL:{\"invite\":{\"$gt\":0}}"`
	IndexPriority = "PRIORITY"
)

type Index struct {
	sch     *Schema
	Name    string
	Unique  bool //唯一
	Sparse  bool //稀疏索引
	Partial []string
	Fields  []*IndexField
}

type IndexField struct {
	Sort     string   // DESC, ASC
	Name     string   // index name
	DBName   []string // a.b.c
	Unique   bool     //唯一
	Sparse   bool     //稀疏索引：仅包含具有索引字段的文档，节省存储空间，适用于字段存在性较少的场景
	Partial  string   //部分索引：基于指定的过滤条件创建索引，只包含满足条件的文档，提高查询性能并减少存储开销，格式如：{"rating":{"$gt":5}}
	Priority int      //排序字段之间的排序ASC
}

func (this *IndexField) GetDBName() string {
	return strings.Join(this.DBName, ".")
}

type IndexPartialParse func(*Schema, []string) (any, error)

func (this *Index) Build(whereParser ...IndexPartialParse) (index *mongo.IndexModel, err error) {
	index = &mongo.IndexModel{}
	var keys []bson.E
	for _, field := range this.Fields {
		k := field.GetDBName()
		v := 1
		if strings.ToUpper(field.Sort) == "DESC" {
			v = -1
		}
		keys = append(keys, bson.E{Key: k, Value: v})
	}
	//fmt.Printf("index:%+v\n\n\n", index)
	index.Keys = keys
	index.Options = options.Index()
	index.Options.SetName(this.Name)
	if this.Unique {
		index.Options.SetUnique(true)
	}
	if this.Sparse {
		index.Options.SetSparse(true)
	}
	if len(this.Partial) > 0 {
		var parser IndexPartialParse
		if len(whereParser) > 0 {
			parser = whereParser[0]
		} else {
			parser = this.partialParser
		}
		var q any
		if q, err = parser(this.sch, this.Partial); err != nil {
			return nil, err
		} else {
			index.Options.SetPartialFilterExpression(q)
		}
	}
	return
}

func (this *Index) partialParser(_ *Schema, s []string) (any, error) {
	q := bson.M{}
	for _, partial := range s {
		var filter bson.M
		if err := json.Unmarshal([]byte(partial), &filter); err != nil {
			return nil, err
		}
		for k, v := range filter {
			q[k] = v
		}
	}
	return q, nil
}

func (schema *Schema) LookIndex(name string) *Index {
	if schema != nil {
		indexes := schema.ParseIndexes()
		for _, index := range indexes {
			if index.Name == name {
				return index
			}
			for _, field := range index.Fields {
				if field.Name == name {
					return index
				}
			}
		}
	}
	return nil
}

// ParseIndexes parse schema indexes
func (schema *Schema) ParseIndexes() map[string]*Index {
	indexes := map[string]*Index{}
	for _, field := range schema.Fields {
		for _, indexField := range schema.parseFieldIndexes(field, "") {
			idx := indexes[indexField.Name]
			if idx == nil {
				idx = &Index{sch: schema, Name: indexField.Name}
				indexes[indexField.Name] = idx
			}
			idx.Fields = append(idx.Fields, indexField)
			if indexField.Unique {
				idx.Unique = indexField.Unique
			}
			if indexField.Sparse {
				idx.Sparse = indexField.Sparse
			}
			if indexField.Partial != "" {
				idx.Partial = append(idx.Partial, indexField.Partial)
			}
		}
	}

	for _, index := range indexes {
		sort.Slice(index.Fields, func(i, j int) bool {
			return index.Fields[i].Priority < index.Fields[j].Priority
		})
	}

	return indexes
}

func (schema *Schema) parseTagIndex(table, field, value string) *IndexField {
	var (
		name     string
		settings = ParseTagSetting(value, ",")
	)

	name = settings[IndexName]
	if name == "" {
		name = strings.Join([]string{"", "idx", table, field}, "_")
	}

	priority, _ := strconv.Atoi(settings[IndexPriority])
	indexField := &IndexField{Name: name, Priority: priority}
	indexField.Sort = settings[IndexSort]
	indexField.DBName = []string{field}

	if _, ok := settings[IndexUnique]; ok {
		indexField.Unique = true
	}
	if _, ok := settings[IndexSparse]; ok {
		indexField.Sparse = true
	}
	if s, ok := settings[IndexPartial]; ok {
		indexField.Partial = s
	}

	return indexField
}

func (schema *Schema) parseFieldIndexes(field *Field, table string) (indexes []*IndexField) {
	if table == "" {
		table = schema.Table
	}
	db := field.DBName()
	if db == "" {
		return
	}

	indexTag, ok := field.StructField.Tag.Lookup(IndexTag)
	if !ok {
		return
	}
	if !strings.Contains(indexTag, ";") {
		indexes = append(indexes, schema.parseTagIndex(table, db, indexTag))
	} else {
		for _, value := range strings.Split(indexTag, ";") {
			if value != "" {
				indexes = append(indexes, schema.parseTagIndex(table, db, value))
			}
		}
	}
	return
}
