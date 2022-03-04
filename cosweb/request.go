package cosweb

import (
	"fmt"
	"github.com/hwcer/cosgo/library/logger"
	"net/url"
)

type RequestDataType int
type RequestDataTypeMap []RequestDataType

const (
	RequestDataTypeParam  RequestDataType = iota //params
	RequestDataTypeForm                          //POST FORM
	RequestDataTypeBody                          //body json and xml....
	RequestDataTypeQuery                         //GET
	RequestDataTypeCookie                        //COOKIES
	RequestDataTypeHeader                        //HEADER
)

func (r RequestDataTypeMap) IndexOf(v RequestDataType) int {
	for i, t := range r {
		if t == v {
			return i
		}
	}
	return -1
}

func (r *RequestDataTypeMap) Add(keys ...RequestDataType) {
	for _, k := range keys {
		if r.IndexOf(k) < 0 {
			*r = append(*r, k)
		}
	}
}
func (r *RequestDataTypeMap) Reset(keys ...RequestDataType) {
	*r = keys
}

//默认获取数据的顺序
var defaultRequestDataType = RequestDataTypeMap{RequestDataTypeParam, RequestDataTypeQuery, RequestDataTypeBody, RequestDataTypeCookie}

func getDataFromRequest(c *Context, key string, dataType RequestDataType) (string, bool) {
	switch dataType {
	case RequestDataTypeParam:
		v, ok := c.params[key]
		return v, ok
	case RequestDataTypeQuery:
		return getQueryValue(c, key)
	case RequestDataTypeBody:
		return getBodyValue(c, key)
	case RequestDataTypeForm:
		if v := c.Request.FormValue(key); v != "" {
			return v, true
		}
	case RequestDataTypeCookie:
		if val, err := c.Request.Cookie(key); err == nil && val.Value != "" {
			return val.Value, true
		}
	case RequestDataTypeHeader:
		if v := c.Request.Header.Get(key); v != "" {
			return v, true
		}
	}
	return "", false
}

func getBodyValue(c *Context, key string) (v string, ok bool) {
	body, err := c.Body()
	if  err != nil {
		return
	}
	var d interface{}
	if d, ok = body[key]; !ok {
		return
	}
	switch d.(type) {
	case string:
		v = d.(string)
	default:
		v = fmt.Sprintf("%v", d)
	}
	return
}

func getQueryValue(c *Context, key string) (v string, ok bool) {
	if c.query == nil {
		var err error
		if c.query, err = url.ParseQuery(c.Request.URL.RawQuery); err != nil {
			logger.Debug("url.ParseQuery Err:%v", err)
			c.query = make(url.Values)
		}

	}
	if ok = c.query.Has(key); ok {
		v = c.query.Get(key)
	}
	return
}
