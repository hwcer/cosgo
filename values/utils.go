package values

func New(data interface{}) *Message {
	return &Message{Data: data}
}

func Error(code int, err interface{}, args ...interface{}) (r *Message) {
	r = &Message{}
	r.SetCode(code, err, args...)
	return
}

func Parse(v interface{}) (r *Message) {
	var ok bool
	if r, ok = v.(*Message); ok {
		return r
	} else if _, ok = v.(error); ok {
		return Error(0, v)
	} else {
		return New(v)
	}
}
