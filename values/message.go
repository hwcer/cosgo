package values

import (
	"fmt"
)

const MessageErrorCodeDefault int = 9999

type Message struct {
	Code int   `json:"code"`
	Data Bytes `json:"data"`
}

func (this *Message) Parse(v interface{}) *Message {
	if r, ok := v.(*Message); ok {
		return r
	}
	if _, ok := v.(error); ok {
		this.Errorf(0, v)
	} else {
		if err := this.Marshal(v); err != nil {
			this.Errorf(0, err)
		}
	}
	return this
}
func (this *Message) String() string {
	if this.Code == 0 {
		return string(this.Data)
	}
	var r string
	if err := this.Data.Unmarshal(&r); err == nil {
		return r
	} else {
		return err.Error()
	}
}

func (this *Message) Error() string {
	return fmt.Sprintf("[%v]%v", this.Code, this.String())
}

// Errorf 格式化一个错误,必定产生错误码
func (this *Message) Errorf(code int, format interface{}, args ...interface{}) {
	if code == 0 {
		this.Code = MessageErrorCodeDefault
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
	_ = this.Data.Marshal(data)
}

func (this *Message) Marshal(v interface{}) error {
	return this.Data.Marshal(v)
}

// Unmarshal 如果是一个错误信息 应当单独处理
func (this *Message) Unmarshal(i interface{}) error {
	if this.Code != 0 {
		return this
	}
	return this.Data.Unmarshal(i)
}

func Parse(v interface{}) *Message {
	if v == nil {
		return &Message{}
	}
	if r, ok := v.(*Message); ok {
		return r
	}
	r := &Message{}
	r = r.Parse(v)
	return r
}

func Errorf(code int, err interface{}, args ...interface{}) (r *Message) {
	r = &Message{}
	r.Errorf(code, err, args...)
	return
}

func NewError(code int, err interface{}, args ...interface{}) (r *Message) {
	return Errorf(code, err, args...)
}

func NewMessage(v interface{}) *Message {
	return Parse(v)
}
