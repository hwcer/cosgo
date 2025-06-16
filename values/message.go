package values

import (
	"fmt"
)

const MessageErrorCodeDefault int32 = 9999

type Message struct {
	Code int32 `json:"code"`
	Data any   `json:"data"`
}

func (this *Message) Parse(v any) *Message {
	switch v.(type) {
	case error:
		this.Errorf(0, v)
	default:
		this.Data = v
	}
	return this
}
func (this *Message) String() string {
	switch v := this.Data.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", this.Data)
	}
}

func (this *Message) Error() string {
	return this.String()
}

// Errorf 格式化一个错误,必定产生错误码
func (this *Message) Errorf(code int32, format any, args ...any) {
	if code == 0 {
		this.Code = MessageErrorCodeDefault
	} else {
		this.Code = code
	}
	this.Data = Sprintf(format, args...)
}

func Parse(v any) *Message {
	switch d := v.(type) {
	case *Message:
		return d
	case Message:
		return &d
	default:
		r := &Message{}
		return r.Parse(v)
	}
}

func Error(err any) (r *Message) {
	r = &Message{}
	r.Errorf(0, err)
	return
}
func Errorf(code int32, format any, args ...any) *Message {
	if r, ok := format.(*Message); ok {
		if code != 0 {
			r.Code = code
		}
		return r
	}
	r := &Message{}
	r.Errorf(code, format, args...)
	return r
}
