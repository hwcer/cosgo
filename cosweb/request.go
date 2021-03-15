package cosweb

const (
	RequestDataTypeParam  int = iota //params
	RequestDataTypeBody              //POST FORM
	RequestDataTypeQuery             //GET
	RequestDataTypeCookie            //COOKIES
	RequestDataTypeHeader            //HEADER
)

//默认获取数据的顺序
var defaultGetRequestDataType []int

func init() {
	defaultGetRequestDataType = append(defaultGetRequestDataType, RequestDataTypeParam, RequestDataTypeQuery, RequestDataTypeBody, RequestDataTypeCookie)
}

func GetDataFromRequest(c *Context, key string, dataType int) (string, bool) {
	switch dataType {
	case RequestDataTypeParam:
		v, ok := c.params[key]
		return v, ok
	case RequestDataTypeBody, RequestDataTypeQuery:
		v := c.Request.FormValue(key)
		return v, true
	case RequestDataTypeCookie:
		v, err := c.Request.Cookie(key)
		if err == nil {
			return v.Value, true
		} else {
			return "", false
		}
	case RequestDataTypeHeader:
		v := c.Request.Header.Get(key)
		return v, true
	default:
		return "", false
	}
}
