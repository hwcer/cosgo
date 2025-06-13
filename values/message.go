package values

import (
	"encoding/json"
	"errors"
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

func (this *Message) Marshal(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	this.Data = b
	return nil
}

// Unmarshal 如果是一个错误信息 应当单独处理
func (this *Message) Unmarshal(i interface{}) (err error) {
	if this.Code != 0 {
		return errors.New(this.String())
	}
	switch v := this.Data.(type) {
	case []byte:
		if len(v) > 0 {
			err = json.Unmarshal(v, i)
		}
	case string:
		if b := []byte(v); len(b) > 0 {
			err = json.Unmarshal(b, i)
		}
	default:
		err = errors.New("unsupported type")
	}
	return
}

type messageWithNet struct {
	Code int32 `json:"code"`
	Data Bytes `json:"data"`
}

func (this *Message) MarshalJSON() ([]byte, error) {
	v := &messageWithNet{Code: this.Code}
	switch i := this.Data.(type) {
	case []byte:
		v.Data = i
	default:
		if b, err := json.Marshal(this.Data); err == nil {
			return nil, err
		} else {
			v.Data = b
		}
	}
	return json.Marshal(v)
}

func (this *Message) UnmarshalJSON(b []byte) error {
	v := &messageWithNet{}
	if err := json.Unmarshal(b, v); err != nil {
		return err
	}
	this.Code = v.Code
	this.Data = []byte(v.Data)
	return nil
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
