package xlsx

import (
	"fmt"
	"github.com/hwcer/cosgo/logger"
	"strings"
	"text/template"
)

var tpl *template.Template

const TemplateTitle = `syntax = "proto3";
option go_package = "./;<% .Package %>";
`

const TemplateDummy = `message <%.Name%>{ <%range .Fields%>
	<%.Type%> <%.Name%> = <%.ProtoIndex%>;<%end%>
}
`

const TemplateMessage = `
<%-  $suffix:=.Suffix %>
<%- range .Sheets%>
message <%.ProtoName%><%$suffix%>{
	<%- range .Fields %>
	<%ProtoRequire .ProtoRequire%><%.ProtoType%> <%.Name%> = <%.ProtoIndex%>; //<% .ProtoDesc%><%end%>
}
<%- if IsArray .ExportType %>
message <%.ProtoName%><%$suffix%>Array{
	repeated <%.ProtoName%><%$suffix%> Coll = 1;
}
<%- end%>
<%- end%>
`

// TemplateSummary 输出一个总表
const TemplateSummary = `
message <%.Name%>{
<%- range .Sheets%>
	<%SummaryType .%> <%.ProtoName%>=<%.ProtoIndex%>;
<%- end %>
}
`

func init() {
	tpl = template.New("")
	tpl.Funcs(template.FuncMap{
		"IsArray":      TemplateIsArray,
		"SummaryType":  TemplateSummaryType,
		"ProtoRequire": TemplateProtoRequire,
	})
	tpl.Delims("<%", "%>")
}

func TemplateIsArray(t ExportType) bool {
	return t == ExportTypeARR
}

func TemplateProtoRequire(t ProtoRequire) string {
	switch t {
	case FieldTypeArray, FieldTypeArrObj:
		return "repeated "
	default:
		return ""
	}
}

func TemplateSummaryType(sheet *Message) string {
	primary := sheet.Fields[0]
	var t string
	switch sheet.ExportType {
	case ExportTypeKVS:
		value := sheet.Fields[1]
		t = value.ProtoType
	case ExportTypeARR:
		t = fmt.Sprintf("%v%vArray", sheet.ProtoName, Config.Suffix)
	default:
		t = fmt.Sprintf("%v%v", sheet.ProtoName, Config.Suffix)
	}
	return fmt.Sprintf("map<%v,%v>", primary.ProtoType, t)
}

func ProtoTitle(builder *strings.Builder) {
	t, err := tpl.Parse(TemplateTitle)
	if err != nil {
		logger.Fatal(err)
	}
	data := &struct {
		Package string
	}{
		Package: Config.Package,
	}
	err = t.Execute(builder, data)
	if err != nil {
		logger.Fatal(err)
	}
	return
}
func ProtoDummy(dummy *Dummy, builder *strings.Builder) {
	t, err := tpl.Parse(TemplateDummy)
	if err != nil {
		logger.Fatal(err)
	}
	err = t.Execute(builder, dummy)
	if err != nil {
		logger.Fatal(err)
	}
	return
}

func ProtoMessage(sheets []*Message, builder *strings.Builder) {
	t, err := tpl.Parse(TemplateMessage)
	if err != nil {
		logger.Fatal(err)
	}
	data := &struct {
		Suffix string
		Sheets []*Message
	}{
		Suffix: Config.Suffix,
		Sheets: sheets,
	}

	err = t.Execute(builder, data)
	if err != nil {
		logger.Fatal(err)
	}

	t, err = tpl.Parse(TemplateSummary)
	if err != nil {
		logger.Fatal(err)
	}
	//输出总表
	data2 := &struct {
		Name   string
		Sheets []*Message
	}{
		Name:   Config.Summary,
		Sheets: sheets,
	}
	err = t.Execute(builder, data2)
	if err != nil {
		logger.Fatal(err)
	}

	return
}
