package message

func New() *Message {
	return &Message{}
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
