package xlsx

import (
	"fmt"
	"sort"
	"strings"
)

type DummyField struct {
	Type  string
	Name  string
	label string //NameType
	Index int    //最终索引ProtoIndex
}

func NewDummy() *Dummy {
	return &Dummy{}
}

type Dummy struct {
	Name   string
	Label  string
	Sheets []int
	Fields []*DummyField
}

func (this *Dummy) Get(name string) *DummyField {
	for _, v := range this.Fields {
		if v.Name == name {
			return v
		}
	}
	return nil
}

func (this *Dummy) AddField(name, fieldType string, sheetIndex int) error {
	name = strings.TrimSpace(name)
	fieldType = TPLProtoType(fieldType)
	if field := this.Get(name); field != nil {
		return fmt.Errorf("字段名重复:%v", name)
	}
	field := &DummyField{
		Type: fieldType,
		Name: name,
	}
	field.label = fmt.Sprintf("%v%v", FirstUpper(field.Name), FirstUpper(field.Type))
	this.Fields = append(this.Fields, field)
	this.Sheets = append(this.Sheets, sheetIndex)
	return nil
}
func (this *Dummy) AddSheet(i int) {
	this.Sheets = append(this.Sheets, i)
}

// Compile 编译并返回全局唯一名字(标记)
func (this *Dummy) Compile() (string, error) {
	if this.Name != "" {
		return this.Name, nil
	}
	sort.Slice(this.Fields, this.less)
	var arr []string
	for i, v := range this.Fields {
		v.Index = i + 1
		arr = append(arr, v.label)
	}
	this.Label = strings.Join(arr, "")
	this.Name = this.Label
	return this.Label, nil
}

func (this *Dummy) less(i, j int) bool {
	return this.Fields[i].label < this.Fields[j].label
}
