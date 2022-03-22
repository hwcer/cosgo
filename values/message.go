package values

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

func (this *Message) SetCode(code int, err interface{}, args ...interface{}) {
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
}

func (this *Message) SetData(v interface{}) {
	this.Data = v
}
