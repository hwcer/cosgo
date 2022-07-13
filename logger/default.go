package logger

import "fmt"

const DefTimeFormat = "2006-01-02 15:04:05 -0700" // 日志输出默认格式

var errorLevelInvalid = fmt.Errorf("无效的日志等级")
var DefaultLogger *Logger
var DefaultAdapter Adapter

func init() {
	DefaultLogger = New(3)
	DefaultAdapter = NewConsoleAdapter()
	_ = DefaultLogger.Adapter(DefaultAdapter)
}

func Alert(f interface{}, v ...interface{}) {
	DefaultLogger.Alert(f, v...)
}

func Fatal(f interface{}, v ...interface{}) {
	DefaultLogger.Fatal(f, v...)
}

func Error(f interface{}, v ...interface{}) {
	DefaultLogger.Error(f, v...)
}

func Warn(f interface{}, v ...interface{}) {
	DefaultLogger.Warn(f, v...)
}

func Info(f interface{}, v ...interface{}) {
	DefaultLogger.Info(f, v...)
}

func Debug(f interface{}, v ...interface{}) {
	DefaultLogger.Debug(f, v...)
}

func Trace(f interface{}, v ...interface{}) {
	DefaultLogger.Trace(f, v...)
}
