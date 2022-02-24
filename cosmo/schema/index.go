package schema

import (
	"sort"
	"strconv"
	"strings"
)

type Index struct {
	Name   string
	Unique bool //唯一
	Sparse bool //稀疏索引
	Fields []IndexOption
}

type IndexOption struct {
	*Field
	Sort     string // DESC, ASC
	priority int    //排序
}

// ParseIndexes parse schema indexes
func (schema *Schema) ParseIndexes() map[string]Index {
	indexes := map[string]Index{}
	for _, field := range schema.Fields {
		if field.TagSettings["INDEX"] != "" || field.TagSettings["UNIQUEINDEX"] != "" {
			for _, index := range parseFieldIndexes(field) {
				idx := indexes[index.Name]
				idx.Name = index.Name
				if index.Unique {
					idx.Unique = index.Unique
				}
				if index.Sparse {
					idx.Sparse = index.Sparse
				}
				idx.Fields = append(idx.Fields, index.Fields...)
				sort.Slice(idx.Fields, func(i, j int) bool {
					return idx.Fields[i].priority < idx.Fields[j].priority
				})

				indexes[index.Name] = idx
			}
		}
	}

	return indexes
}

func (schema *Schema) LookIndex(name string) *Index {
	if schema != nil {
		indexes := schema.ParseIndexes()
		for _, index := range indexes {
			if index.Name == name {
				return &index
			}
			for _, field := range index.Fields {
				if field.Name == name {
					return &index
				}
			}
		}
	}
	return nil
}

func parseFieldIndexes(field *Field) (indexes []Index) {
	for _, value := range strings.Split(field.Tag.Get("gorm"), ";") {
		if value != "" {
			v := strings.Split(value, ":")
			k := strings.TrimSpace(strings.ToUpper(v[0]))
			if k == "INDEX" || k == "UNIQUEINDEX" {
				var (
					name     string
					tag      = strings.Join(v[1:], ":")
					idx      = strings.Index(tag, ",")
					settings = ParseTagSetting(tag, ",")
					//length, _ = strconv.Atoi(settings["LENGTH"])
				)

				if idx == -1 {
					idx = len(tag)
				}

				if idx != -1 {
					name = tag[0:idx]
				}

				if name == "" {
					name = field.Schema.namer.IndexName(field.Schema.Table, field.Name)
				}

				//if (k == "UNIQUEINDEX") || settings["UNIQUE"] != "" {
				//	settings["CLASS"] = "UNIQUE"
				//}
				//if settings["Sparse"] != "" {
				//	settings["CLASS"] = "UNIQUE"
				//}

				priority, err := strconv.Atoi(settings["PRIORITY"])
				if err != nil {
					priority = 10
				}
				index := Index{
					Name:   name,
					Fields: []IndexOption{{Field: field, Sort: settings["SORT"], priority: priority}},
				}
				if _, ok := settings["UNIQUE"]; ok {
					index.Unique = true
				}
				if _, ok := settings["SPARSE"]; ok {
					index.Sparse = true
				}
				indexes = append(indexes, index)
			}
		}
	}

	return
}
