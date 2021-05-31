package cosweb

var Options = struct {
	SessionName   string
	SessionSecret string
	SessionMethod RequestDataTypeMap
}{
	SessionName:   "_cosweb_cookie_key",
	SessionMethod: RequestDataTypeMap{RequestDataTypeCookie, RequestDataTypeHeader, RequestDataTypeParam},
}
