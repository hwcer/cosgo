package cosweb

type RequestDataType uint8

const (
	RequestDataTypePath   RequestDataType = iota //params
	RequestDataTypeBody                          //POST，JSON,FORM
	RequestDataTypeQuery                         //GET
	RequestDataTypeCookie                        //COOKIES
)

//默认获取数据的顺序
var defaultGetRequestDataType []RequestDataType

func init() {
	defaultGetRequestDataType = append(defaultGetRequestDataType, RequestDataTypePath, RequestDataTypeQuery, RequestDataTypeBody, RequestDataTypeCookie)
}

func GetDataFromRequest(c *Context, key string, dataType RequestDataType) (string, bool) {
	switch dataType {
	case RequestDataTypePath:
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
			//LOGGER
			return "", false
		}
	default:
		return "", false
	}
}
