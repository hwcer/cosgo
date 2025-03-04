package binder

import "strings"

const (
	HeaderAccept      = "Accept"
	HeaderContentType = "Content-Type"
)

var mimeTypes = map[string]string{} // JSON --> application/json
var mimeNames = map[string]string{} //  application/json  --> JSON

func GetMimeName(t string) string {
	s := strings.ToLower(t)
	if v, ok := mimeNames[s]; ok {
		return v
	}
	return t
}

func GetMimeType(t string) string {
	s := strings.ToUpper(t)
	if v, ok := mimeTypes[s]; ok {
		return v
	}
	return t
}

func SetMimeType(name string, typ string) {
	name = strings.ToUpper(name)
	mimeTypes[name] = typ
	mimeNames[typ] = name
}

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON              string = "application/json"
	MIMEXML                      = "application/xml"
	MIMEXML2                     = "text/xml"
	MIMEPOSTForm                 = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm        = "multipart/form-data"
	MIMEPROTOBUF                 = "application/x-protobuf"
	MIMEMSGPACK                  = "application/x-msgpack"
	MIMEMSGPACK2                 = "application/msgpack"
	MIMEYAML                     = "application/x-yaml"
)

func init() {
	SetMimeType("JSON", MIMEJSON)
	SetMimeType("XML", MIMEXML)
	SetMimeType("XML2", MIMEXML2)
	SetMimeType("PROTOBUF", MIMEPROTOBUF)
	SetMimeType("MSGPACK", MIMEMSGPACK)
	SetMimeType("MSGPACK2", MIMEMSGPACK2)
	SetMimeType("YAML", MIMEYAML)
}
