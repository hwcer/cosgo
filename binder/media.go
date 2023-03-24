package binder

const ContentType = "Content-Type"

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
