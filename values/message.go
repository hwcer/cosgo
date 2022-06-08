package values

import (
	"fmt"
)

var DefaultErrorCode int = 9999

type Message struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func (this *Message) Parse(v interface{}) *Message {
	if r, ok := v.(*Message); ok {
		this.Code = r.Code
		this.Data = r.Data
	} else if _, ok2 := v.(error); ok2 {
		this.Errorf(0, v)
	} else {
		this.SetData(v)
	}
	return this
}

func (this *Message) Error() string {
	return fmt.Sprintf("%v,code:%v", this.Data, this.Code)
}

func (this *Message) Errorf(code int, err interface{}, args ...interface{}) *Message {
	if code == 0 {
		this.Code = DefaultErrorCode
	} else {
		this.Code = code
	}
	msg, ok := err.(string)
	if !ok {
		msg = fmt.Sprintf("%v", err)
	}
	if len(args) > 0 {
		msg = fmt.Sprintf(msg, args...)
	}
	this.Data = msg
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
