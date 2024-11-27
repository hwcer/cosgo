package schema

import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"sort"
	"strconv"
	"strings"
)

const (
	IndexTag      = "index"
	IndexName     = "NAME"
	IndexSort     = "SORT"
	IndexUnique   = "UNIQUE"
	IndexSparse   = "SPARSE"
	IndexPriority = "PRIORITY"
)

type Index struct {
	Name   string
	Unique bool //唯一
	Sparse bool //稀疏索引
	Fields []*IndexField
}

type IndexField struct {
	//*Field
	Sort     string   // DESC, ASC
	Name     string   // index name
	DBName   []string // a.b.c
	Unique   bool     //唯一
	Sparse   bool     //稀疏索引
	Priority int      //排序字段之间的排序ASC
}

func (this *IndexField) GetDBName() string {
	return strings.Join(this.DBName, ".")
}

func (this *Index) Build() (index mongo.IndexModel) {
	index = mongo.IndexModel{}
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
	return index
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
				idx = &Index{Name: indexField.Name}
				indexes[indexField.Name] = idx
			}
			idx.Fields = append(idx.Fields, indexField)
			if indexField.Unique {
				idx.Unique = indexField.Unique
			}
			if indexField.Sparse {
				idx.Sparse = indexField.Sparse
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
	return indexField
}

func (schema *Schema) parseFieldIndexes(field *Field, table string) (indexes []*IndexField) {
	if field.DBName == "" || field.DBName == "-" {
		return
	}
	if table == "" {
		table = schema.Table
	}
	indexTag, ok := field.StructField.Tag.Lookup(IndexTag)
	if !ok {
		return
	}
	if !strings.Contains(indexTag, ";") {
		indexes = append(indexes, schema.parseTagIndex(table, field.DBName, indexTag))
	} else {
		for _, value := range strings.Split(indexTag, ";") {
			if value != "" {
				indexes = append(indexes, schema.parseTagIndex(table, field.DBName, value))
			}
		}
	}
	//递归子对象
	//if field.IndirectFieldType.Kind() != reflect.Struct {
	//	return
	//}
	//fieldValue := reflect.New(field.IndirectFieldType)
	//fieldSchema, err := GetOrParse(fieldValue.Interface(), schema.options)
	//if err != nil {
	//	fmt.Printf("Schema parseFieldIndexes:%v", err)
	//	return
	//}
	//fieldTable := strings.Join([]string{table, field.DBName}, "_")
	//for _, v := range fieldSchema.fieldsPrivate {
	//	for _, index := range fieldSchema.parseFieldIndexes(v, fieldTable) {
	//		index.DBName = append([]string{field.DBName}, index.DBName...)
	//		indexes = append(indexes, index)
	//	}
	//}

	return
}
