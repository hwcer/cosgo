package binder

import "mime"

// Content-Type MIME of the most common data formats.
const (
	MIMEJSON              string = "application/json"
	MIMEHTML                     = "text/html"
	MIMEXML                      = "application/xml"
	MIMEXML2                     = "text/xml"
	MIMEPlain                    = "text/plain"
	MIMEPOSTForm                 = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm        = "multipart/form-data"
	MIMEPROTOBUF                 = "application/x-protobuf"
	MIMEMSGPACK                  = "application/x-msgpack"
	MIMEMSGPACK2                 = "application/msgpack"
	MIMEYAML                     = "application/x-yaml"
)

var mediaTypeDict = make(map[string]EncodingType)

func init() {
	mediaTypeDict[MIMEXML] = EncodingTypeXml
	mediaTypeDict[MIMEXML2] = EncodingTypeXml
	mediaTypeDict[MIMEJSON] = EncodingTypeJson
	mediaTypeDict[MIMEPOSTForm] = EncodingTypeUrlEncoded
	mediaTypeDict[MIMEPROTOBUF] = EncodingTypeProtoBuf
	mediaTypeDict[MIMEYAML] = EncodingTypeYaml
}

// ParseMediaType MediaType to binding type
func ParseMediaType(name string) (t EncodingType, err error) {
	var ct string
	ct, _, err = mime.ParseMediaType(name)
	if err == nil {
		t = mediaTypeDict[ct]
	}
	return
}

// ParseMediaHandle MediaType to binding handle
func ParseMediaHandle(name string) (r Interface) {
	if t, _ := ParseMediaType(name); t > 0 {
		r = binderMap[t]
	}
	return
}
