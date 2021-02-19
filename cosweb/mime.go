package cosweb

const charsetUTF8 = "charset=UTF-8"

// MIME types
const (
	MIMETextHTML  = "text/html"
	MIMETextPlain = "text/plain"

	MIMETextXML         = "text/xml"
	MIMEApplicationJS   = "application/javascript"
	MIMEApplicationXML  = "application/xml"
	MIMEApplicationJSON = "application/json"

	MIMEApplicationProtobuf = "application/protobuf"
	MIMEApplicationMsgpack  = "application/msgpack"

	MIMEOctetStream = "application/octet-stream"

	MIMEMultipartForm   = "multipart/form-data"
	MIMEApplicationForm = "application/x-www-form-urlencoded"
)

//MIME UTF8 types
const (
	MIMETextHTMLCharsetUTF8  = MIMETextHTML + "; " + charsetUTF8
	MIMETextPlainCharsetUTF8 = MIMETextPlain + "; " + charsetUTF8

	MIMETextXMLCharsetUTF8 = MIMETextXML + "; " + charsetUTF8

	MIMEApplicationJSCharsetUTF8   = MIMEApplicationJS + "; " + charsetUTF8
	MIMEApplicationXMLCharsetUTF8  = MIMEApplicationXML + "; " + charsetUTF8
	MIMEApplicationJSONCharsetUTF8 = MIMEApplicationJSON + "; " + charsetUTF8
)
