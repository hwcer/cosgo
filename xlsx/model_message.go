package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo/logger"
	"github.com/hwcer/cosgo/utils"
	"github.com/tealeg/xlsx/v3"
	"strings"
)

var ignoreFiles []string
var globalObjects = GlobalDummy{}

type ExportType int8
type GlobalDummy map[string]*Dummy

const (
	ExportTypeMap ExportType = iota //默认
	ExportTypeKVS
	ExportTypeARR
)

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

type Message struct {
	Fields     []*Field       //字段列表
	FileName   string         //文件名
	SheetName  string         //表格名称
	SheetType  map[int]string //表格属性
	SheetDesc  map[int]string //表格描述
	LowerName  string         // 小写的表名字， 保证唯一
	ProtoName  string         // protoName 是pb.go中文件的名字，
	ProtoIndex int            //总表中的序号
	SheetRows  *xlsx.Sheet    //sheets
	SheetSkip  int            //数据表中数据部分需要跳过的行数
	ExportType ExportType     //输出类型,kv arr map
}

//const RowId = "id"

type rowArr struct {
	Coll []any
}

func (this *Message) Values() (any, []error) {
	r := map[string]any{}
	var errs []error
	var emptyCell []int
	max := this.SheetRows.MaxRow
	for i := this.SheetSkip + 1; i <= max; i++ {
		row, err := this.SheetRows.Row(i)
		if err != nil {
			logger.Info("%v,err:%v", i, err)
		}
		id := strings.TrimSpace(row.GetCell(0).Value)
		if utils.Empty(id) {
			emptyCell = append(emptyCell, row.GetCoordinate()+1)
			continue
		}
		//KV 模式直接定位 0,1 列
		if this.ExportType == ExportTypeKVS {
			var data any
			field := this.Fields[1]
			if data, err = field.Value(row); err == nil {
				r[id] = data
			} else {
				errs = append(errs, fmt.Errorf("解析错误:%v第%v行,%v", this.ProtoName, row.GetCoordinate()+1, err))
			}
			continue
		}
		//MAP ARRAY
		val, err := this.Value(row)
		if err != nil {
			errs = append(errs, fmt.Errorf("解析错误:%v第%v行,%v", this.ProtoName, row.GetCoordinate()+1, err))
			continue
		}
		if this.ExportType == ExportTypeARR {
			if d, ok := r[id]; !ok {
				d2 := &rowArr{}
				d2.Coll = append(d2.Coll, val)
				r[id] = d2
			} else {
				d2, _ := d.(*rowArr)
				d2.Coll = append(d2.Coll, val)
			}
		} else {
			r[id] = val
		}
	}

	if len(emptyCell) > 10 {
		logger.Info("%v共%v行ID为空已经忽略:%v", this.ProtoName, len(emptyCell), emptyCell)
	}
	return r, errs
}

func (this *Message) Value(row *xlsx.Row) (map[string]any, error) {
	r := map[string]any{}
	for _, field := range this.Fields {
		v, e := field.Value(row)
		if e != nil {
			return nil, e
		} else {
			r[field.Name] = v
		}
	}
	return r, nil
}

// GlobalObjectsProtoName 通过ProtoName生成对象
func (this *Message) GlobalObjectsProtoName() {
	for _, field := range this.Fields {
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
func (this *Message) GlobalObjectsAutoName() {
	for _, field := range this.Fields {
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

func buildGlobalObjects(b *strings.Builder, sheets []*Message) {
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
