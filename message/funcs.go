package message

import "fmt"

func New(v interface{}) *Message {
	if r, ok := v.(*Message); ok {
		return r
	}
	var err error
	r := &Message{}
	r, err = r.Parse(v)
	if err != nil {
		fmt.Printf("%v\n", err)
	}
	return r
}

func Parse(v interface{}) *Message {
	return New(v)
}

func Error(err interface{}) (r *Message) {
	r = &Message{}
	_ = r.SetError(0, err)
	return
}

func Errorf(code int, err interface{}, args ...interface{}) (r *Message) {
	r = &Message{}
	_ = r.SetError(code, err, args...)
	return
}
