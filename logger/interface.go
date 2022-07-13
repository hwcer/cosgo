package logger

type Interface interface {
	Fatal(format interface{}, args ...interface{})
	Alert(format interface{}, args ...interface{})
	Error(format interface{}, args ...interface{})
	Warn(format interface{}, args ...interface{})
	Info(format interface{}, args ...interface{})
	Debug(format interface{}, args ...interface{})
	Trace(format interface{}, args ...interface{})
}

type Adapter interface {
	Init() error
	Write(msg *Message, level int) error
	Close()
}
