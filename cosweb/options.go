package cosweb

var Options = struct {
	Debug   bool //DEBUG模式会打印所有路由匹配状态,向客户端输出详细错误信息
	Session sessionOptions
}{
	Session: sessionOptions{
		Name:   "_cosweb_cookie_key",
		Method: RequestDataTypeMap{RequestDataTypeCookie, RequestDataTypeHeader, RequestDataTypeParam},
	},
}
