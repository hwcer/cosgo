package network

import "fmt"

var Logger logger

type logger interface {
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
	Debug(format string, v ...interface{})
}

func SetLogger(log logger)  {
	Logger = log
}

type defaultLogger struct {

}
func (l *defaultLogger) Warn(format string, v ...interface{})  {
	fmt.Printf(format+"\n",v...)
}

func (l *defaultLogger) Error(format string, v ...interface{})  {
	fmt.Printf(format+"\n",v...)
}

func (l *defaultLogger) Debug(format string, v ...interface{})  {
	fmt.Printf(format+"\n",v...)
}