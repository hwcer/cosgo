package message

import (
	"fmt"
	"github.com/hwcer/cosgo/values"
)

var DefaultErrorCode int = 9999

type Message struct {
	Code int          `json:"code"`
	Data values.Bytes `json:"data"`
}

func (this *Message) Parse(v interface{}) *Message {
	if r, ok := v.(*Message); ok {
		return r
	}
	var err error
	if _, ok := v.(error); ok {
		err = this.SetError(0, v)
	} else {
		err = this.SetData(v)
	}
	if err != nil {
		_ = this.SetError(0, err.Error())
	}
	return this
}

func (this *Message) Error() string {
	return fmt.Sprintf("%v,code:%v", this.Data, this.Code)
}

func (this *Message) SetCode(v int) {
	this.Code = v
}
func (this *Message) SetData(v interface{}) error {
	return this.Data.Marshal(v)
}

func (this *Message) SetError(code int, format interface{}, args ...interface{}) (err error) {
	if code == 0 {
		this.Code = DefaultErrorCode
	} else {
		this.Code = code
	}
	var data string
	switch v := format.(type) {
	case string:
		if len(args) > 0 {
			data = fmt.Sprintf(v, args...)
		} else {
			data = v
		}
	default:
		data = fmt.Sprintf("%v", format)
	}
	err = this.Marshal(data)
	return
}

// Marshal 将一个对象放入data
func (this *Message) Marshal(v interface{}) error {
	return this.Data.Marshal(v)
}

// Unmarshal 使用i解析data
func (this *Message) Unmarshal(i interface{}) error {
	return this.Data.Unmarshal(i)
}
