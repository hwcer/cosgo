package xlsx

import (
	"github.com/hwcer/cosgo/logger"
	"strings"
	"text/template"
)

var tpl *template.Template

const TemplateTitle = `syntax = "proto3";
option go_package = "./;{{.Package}}";
`

const TemplateDummy = `message {{.Name}}{ {{range .Fields}}
	{{ProtoType .Type}} {{.Name}} = {{.Index}};{{end}}
}
`

const TemplateMessage = `{{range .}} message {{.ProtoName}}{ {{range .SheetFields}}
	{{ProtoRequire .ProtoRequire}}{{ProtoType .ProtoType}} {{.Name}} = {{.ProtoIndex}};{{end}}
}
{{end}}
`

func TPLProtoType(t string) string {
	switch t {
	case "int", "int32":
		return "int32"
	case "int64":
		return "int64"
	case "str", "string", "text":
		return "string"
	}
	return t
}

func TPLProtoRequire(t ProtoRequire) string {
	switch t {
	case FieldTypeArray, FieldTypeArrObj:
		return "repeated "
	default:
		return ""
	}
}

func init() {
	tpl = template.New("")
	tpl.Funcs(template.FuncMap{
		"ProtoType":    TPLProtoType,
		"ProtoRequire": TPLProtoRequire,
	})
}

func ProtoTitle(builder *strings.Builder) {
	t, err := tpl.Parse(TemplateTitle)
	if err != nil {
		logger.Fatal(err)
	}
	data := &struct {
		Package string
	}{
		Package: "gameData",
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

func ProtoMessage(sheets []*GameSheet, builder *strings.Builder) {
	t, err := tpl.Parse(TemplateMessage)
	if err != nil {
		logger.Fatal(err)
	}

	//data := &struct {
	//	Sheets []*GameSheet
	//}{
	//	Sheets: sheets,
	//}
	err = t.Execute(builder, sheets)
	if err != nil {
		logger.Fatal(err)
	}
	return
}
