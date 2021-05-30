package cosweb

type Options struct {
	Session *Session
}

func NewOptions() *Options {
	return &Options{
		Session: NewSession(),
	}
}
