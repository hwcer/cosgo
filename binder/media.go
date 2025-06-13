package binder

import "strings"

type T struct {
	Id   uint8
	Name string
	Type string
}

var mimeIds = map[uint8]*T{}    //ID索引
var mimeTypes = map[string]*T{} // application/json
var mimeNames = map[string]*T{} // JSON

func SetMimeType(id uint8, name string, typ string) {
	v := &T{}
	v.Id = id
	v.Name = strings.ToUpper(name)
	v.Type = strings.ToLower(typ)
	mimeIds[id] = v
	mimeTypes[v.Type] = v
	mimeNames[v.Name] = v
}

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON     string = "application/json"
	MIMEXML             = "application/xml"
	MIMEXML2            = "text/xml"
	MIMEPOSTForm        = "application/x-www-form-urlencoded"
	MIMEPROTOBUF        = "application/x-protobuf"
	MIMEMSGPACK         = "application/x-msgpack"
	MIMEMSGPACK2        = "application/msgpack"
	MIMEYAML            = "application/x-yaml"
)

func init() {
	SetMimeType(1, "JSON", MIMEJSON)
	SetMimeType(2, "PROTOBUF", MIMEPROTOBUF)
	SetMimeType(3, "MSGPACK", MIMEMSGPACK)
	SetMimeType(4, "XML", MIMEXML)
	SetMimeType(5, "YAML", MIMEYAML)
	SetMimeType(6, "FORM", MIMEPOSTForm)

	SetMimeType(30, "MSGPACK2", MIMEMSGPACK2)
	SetMimeType(40, "XML2", MIMEXML2)

}
