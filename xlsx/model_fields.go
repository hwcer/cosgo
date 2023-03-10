package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo/logger"
	"github.com/tealeg/xlsx/v3"
	"strings"
)

type ProtoRequire int8

const (
	FieldTypeNone ProtoRequire = iota
	FieldTypeArray
	FieldTypeObject
	FieldTypeArrObj
)

type flags []string

func (this *flags) Has(s string) bool {
	for _, v := range *this {
		if s == v {
			return true
		}
	}
	return false
}

// HasAndPop 如果结尾存在s则将s弹出
func (this *flags) HasAndPop(s string) (r string, has bool) {
	r = s
	l := len(*this)
	if l == 0 {
		return
	}
	if v := (*this)[l-1]; v == s {
		has = true
		*this = (*this)[0 : l-1]
	}
	return
}

type Field struct {
	flags        flags
	Name         string
	Index        []int
	Dummy        []*Dummy
	ProtoDesc    string //备注信息
	ProtoName    string //数据集表格中自定义的子对象名字
	ProtoType    string
	ProtoIndex   int
	ProtoRequire ProtoRequire
}

func (this *Field) IsEnd() bool {
	return len(this.flags) == 0
}

// Compile 编译并判断是否合法,必须处理完所有标签和子对象
func (this *Field) Compile() bool {
	if this.Name == "" {
		return false
	}
	if !(this.ProtoRequire == FieldTypeObject || this.ProtoRequire == FieldTypeArrObj) {
		return len(this.flags) == 0
	}
	if len(this.Dummy) == 0 {
		return false
	}

	var label string
	for _, v := range this.Dummy {
		if s, e := v.Compile(); e != nil {
			logger.Fatal("Field Compile:%v", e)
		} else if label == "" {
			label = s
		} else if label != s {
			logger.Fatal("%v 子对象类型不统一:%v -- %v", this.Name, label, s)
		}
	}
	this.ProtoType = label
	return len(this.flags) == 0
}

// 寻找结束符号
func (this *Field) ending(cell *xlsx.Cell, index int, suffix, protoType string) bool {
	if suffix == "" {
		return this.IsEnd()
	}
	var k []string
	var flag flags
	//var end bool
	for _, s := range suffix {
		c := fmt.Sprintf("%c", s)
		if v, has := this.flags.HasAndPop(c); !has {
			k = append(k, v)
		} else {
			flag = append(flag, v)
		}
	}
	if !(this.ProtoRequire == FieldTypeObject || this.ProtoRequire == FieldTypeArrObj) {
		return this.IsEnd()
	}
	if len(k) == 0 {
		logger.Fatal("子对象属性不能为空:%v", this.Name)
	}
	if len(this.Dummy) == 0 {
		logger.Fatal("错误的结束符号:%v", this.Name)
	}
	//开始子属性
	id := strings.Join(k, "")
	dummy := this.Dummy[len(this.Dummy)-1]
	if err := dummy.Add(id, protoType, index); err != nil {
		logger.Fatal(err)
	}
	//fmt.Printf("发现子属性:%v %v\n", this.Name, strings.Join(k, ""))
	//统计子对象属性

	return this.IsEnd()
}

// Parse [{   [[  {  [
func (this *Field) Parse(cell *xlsx.Cell, index int, protoType string) (end bool) {
	if protoType == "" {
		return false
	}
	//this.begin += 1
	this.Index = append(this.Index, index)
	value := cell.Value
	if value == "" {
		return len(this.flags) == 0 //TODO 只有ARRAY允许为空
	}
	//var protoName string
	if i, j := strings.Index(value, "<"), strings.Index(value, ">"); i >= 0 && j >= 0 {
		this.ProtoName = FirstUpper(value[i+1 : j])
		value = value[j+1:]
	}
	//begin := false //不能在同一个单元格内同时开始和结束
	name, suffix := "", ""
	var protoRequire ProtoRequire
	if i := strings.Index(value, "[{"); i >= 0 {
		//begin = true
		name = value[0:i]
		suffix = value[i+2:]
		this.flags = append(this.flags, "]", "}")
		this.Dummy = append(this.Dummy, NewDummy())
		protoRequire = FieldTypeArrObj
	} else if i = strings.Index(value, "["); i >= 0 {
		//begin = true
		name = value[0:i]
		//suffix = value[i:]
		this.flags = append(this.flags, "]")
		protoRequire = FieldTypeArray
	} else if i = strings.Index(value, "{"); i >= 0 {
		//begin = true
		name = value[0:i]
		suffix = value[i+1:]
		this.flags = append(this.flags, "}")
		this.Dummy = append(this.Dummy, NewDummy())
		protoRequire = FieldTypeObject
	} else {
		name = value
		suffix = value
	}
	if len(this.Index) == 1 {
		this.Name = name //第一个名字为准
		this.ProtoIndex = index + 1
		this.ProtoRequire = protoRequire
		if this.ProtoRequire == FieldTypeNone || this.ProtoRequire == FieldTypeArray {
			this.ProtoType = protoType
		}
	}
	if this.ProtoRequire == FieldTypeNone {
		return true
	}
	return this.ending(cell, index, suffix, protoType)
	//if !begin {
	//	return this.ending(cell)
	//}
	//fmt.Printf("发现ID:%v", suffix)
	//return false
}

// Value 根据一行表格获取值
func (this *Field) Value(row *xlsx.Row) (ret any, err error) {
	switch this.ProtoRequire {
	case FieldTypeArray:
		var r []any
		var v any
		for _, i := range this.Index {
			if v, err = FormatValue(row, i, this.ProtoType); err == nil {
				r = append(r, v)
			} else {
				break
			}
		}
		if err == nil {
			ret = r
		}
	case FieldTypeObject:
		ret, err = this.Dummy[0].Value(row)
	case FieldTypeArrObj:
		var r []any
		var v any
		for _, dummy := range this.Dummy {
			if v, err = dummy.Value(row); err == nil {
				r = append(r, v)
			} else {
				break
			}
		}
		if err == nil {
			ret = r
		}
	default:
		ret, err = FormatValue(row, this.Index[0], this.ProtoType)
	}

	if err != nil {
		err = fmt.Errorf("字段名:%v,错误信息:%v", this.Name, err)
	}
	return
}
