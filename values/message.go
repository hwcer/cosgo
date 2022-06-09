package values

import (
	"fmt"
)

var DefaultErrorCode int = 9999

func New(data interface{}) *Message {
	return &Message{Data: data}
}

func Errorf(code int, err interface{}, args ...interface{}) (r *Message) {
	r = &Message{}
	return r.Errorf(code, err, args...)
}

func Parse(v interface{}) *Message {
	if r, ok := v.(*Message); ok {
		return r
	}
	r := &Message{}
	return r.Parse(v)
}

type Message struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func (this *Message) Parse(v interface{}) *Message {
	if r, ok := v.(*Message); ok {
		return r
	} else if _, ok2 := v.(error); ok2 {
		return this.Errorf(0, v)
	} else {
		return this.SetData(v)
	}
}

func (this *Message) Error() string {
	return fmt.Sprintf("%v,code:%v", this.Data, this.Code)
}

func (this *Message) Errorf(code int, format interface{}, args ...interface{}) *Message {
	if code == 0 {
		this.Code = DefaultErrorCode
	} else {
		this.Code = code
	}
	switch format.(type) {
	case string:
		this.Data = fmt.Sprintf(format.(string), args...)
	default:
		this.Data = fmt.Sprintf("%v", format)
	}
	return this
}

func (this *Message) SetCode(v int) *Message {
	this.Code = v
	return this
}

func (this *Message) SetData(v interface{}) *Message {
	this.Data = v
	return this
}
