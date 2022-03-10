package message

import (
	"fmt"
)

var DefaultErrorCode int = 9999

type Message struct {
	Code int         `json:"code"`
	Data interface{} `json:"data"`
}

func (this *Message) Error() string {
	return fmt.Sprintf("%v", this.Data)
}

func New(data interface{}) *Message {
	return &Message{Data: data}
}

func Error(code int, err interface{}, args ...interface{}) (r *Message) {
	if code == 0 {
		code = DefaultErrorCode
	}
	r = &Message{Code: code}
	data := fmt.Sprintf("%v", err)
	if len(args) > 0 {
		data = fmt.Sprintf(data, args...)
	}
	r.Data = data
	return
}

func Serialize(v interface{}) (r *Message) {
	var ok bool
	if r, ok = v.(*Message); ok {
		return r
	} else if _, ok = v.(error); ok {
		return Error(0, v)
	} else {
		return New(v)
	}
}
