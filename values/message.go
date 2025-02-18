package values

import (
	"errors"
	"fmt"
)

const MessageErrorCodeDefault int = 9999

type Message struct {
	Code int   `json:"code"`
	Data Bytes `json:"data"`
}

func (this *Message) Parse(v interface{}) *Message {
	switch d := v.(type) {
	case error:
		this.Format(0, d)
	case []byte:
		this.Data = d
	default:
		if err := this.Marshal(v); err != nil {
			this.Format(0, err)
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
	return this.String()
}

// Format 格式化一个错误,必定产生错误码
func (this *Message) Format(code int, format any, args ...any) {
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
		return errors.New(this.String())
	}
	return this.Data.Unmarshal(i)
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
	r.Format(0, err)
	return
}
func Errorf(code int, format any, args ...any) (r *Message) {
	var ok bool
	if r, ok = format.(*Message); ok {
		if code != 0 {
			r.Code = code
		}
		return r
	}
	r = &Message{}
	r.Format(code, format, args...)
	return
}

//func NewMessage(v interface{}) *Message {
//	return Parse(v)
//}
