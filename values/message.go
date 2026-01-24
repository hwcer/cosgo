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
	if this.Data == nil {
		return ""
	}
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
	return Errorf(0, err)
}
func Errorf(code int32, format any, args ...any) (r *Message) {
	switch v := format.(type) {
	case *Message:
		r = v
	case Message:
		r = &v
	}
	if r != nil {
		if code != 0 {
			r.Code = code
		} else if r.Code == 0 {
			r.Code = MessageErrorCodeDefault
		}
		return r
	}
	r = &Message{}
	r.Errorf(code, format, args...)
	return r
}
