package xlsx

import (
	"github.com/hwcer/cosgo/logger"
	"strings"
)

var ignoreFiles []string
var globalObjects = GlobalDummy{}

type GlobalDummy map[string]*Dummy

// Search 查找可能兼容的对象
func (this *GlobalDummy) Search(d *Dummy) (r string, ok bool) {
	dict := *this
	for k, v := range dict {
		if strings.Contains(d.Label, v.Label) {
			dict[k] = d
			return k, true
		} else if strings.Contains(v.Label, d.Label) {
			return k, true
		}
	}
	return
}

type GameSheet struct {
	//SheetIndex int
	FileName    string
	SheetName   string
	SheetType   map[int]string
	LowerName   string // 小写的表名字， 保证唯一
	ProtoName   string // protoName 是pb.go中文件的名字，
	SheetFields []*Field
	//RowMetas  []RowMeta
	//Rows      [][]GameCell
	//Temp      int
}

// GlobalObjectsProtoName 通过ProtoName生成对象
func (this *GameSheet) GlobalObjectsProtoName() {
	for _, field := range this.SheetFields {
		if (field.ProtoRequire == FieldTypeObject || field.ProtoRequire == FieldTypeArrObj) && field.ProtoName != "" {
			name := field.ProtoName
			dummy := field.Dummy[0]
			if k, ok := globalObjects.Search(dummy); ok {
				field.ProtoType = k
				if name != k {
					logger.Info("冗余的对象名称%v,建议修改成%v;Sheet:%v---->%v", name, k, this.SheetName, this.FileName)
				}
			} else {
				field.ProtoType = name
				globalObjects[name] = dummy
			}
		}
	}
}

// GlobalObjectsAutoName 自动命名
func (this *GameSheet) GlobalObjectsAutoName() {
	for _, field := range this.SheetFields {
		if (field.ProtoRequire == FieldTypeObject || field.ProtoRequire == FieldTypeArrObj) && field.ProtoName == "" {
			dummy := field.Dummy[0]
			if k, ok := globalObjects.Search(dummy); ok {
				field.ProtoType = k
			} else {
				field.ProtoType = dummy.Label
				globalObjects[dummy.Label] = dummy
			}
		}
	}
}

func buildGlobalObjects(b *strings.Builder, sheets []*GameSheet) {
	for _, s := range sheets {
		s.GlobalObjectsProtoName()
	}
	for _, s := range sheets {
		s.GlobalObjectsAutoName()
	}
	for k, dummy := range globalObjects {
		dummy.Name = k
		ProtoDummy(dummy, b)
	}

	globalObjects = map[string]*Dummy{}
}
