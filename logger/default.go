package logger

import (
	"fmt"
)

const DefaultConsoleName = "_default_console_name"

var defaultLogger *Logger

func init() {
	defaultLogger = New(3)
	_ = defaultLogger.SetOutput(DefaultConsoleName, Console)
	//if err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd())); err != nil {
	//	panic(err)
	//}
}

func Fatal(f any, v ...any) {
	defaultLogger.Fatal(f, v...)
}
func Panic(f any, v ...any) {
	defaultLogger.Panic(f, v...)
}
func Error(f any, v ...any) {
	defaultLogger.Error(f, v...)
}

func Alert(f any, v ...any) {
	defaultLogger.Alert(f, v...)
}

func Debug(f any, v ...any) {
	defaultLogger.Debug(f, v...)
}

func Trace(f any, v ...any) {
	defaultLogger.Trace(f, v...)
}

// SetLevel 设置日志输出等级
func SetLevel(level Level) {
	defaultLogger.SetLevel(level)
}

// SetPathTrim 设置日志起始路径
func SetPathTrim(trimPath ...string) {
	defaultLogger.SetPathTrim(trimPath...)
}

func SetCallDepth(depth int) {
	defaultLogger.SetCallDepth(depth)
}

func SetOutput(name string, output Output) error {
	return defaultLogger.SetOutput(name, output)
}
func DelOutput(name string) {
	defaultLogger.DelOutput(name)
}
func Sprintf(format any, args ...any) (text string) {
	switch v := format.(type) {
	case string:
		text = v
	case error:
		text = v.Error()
	default:
		text = fmt.Sprintf("%v", format)
	}
	if len(args) > 0 {
		text = fmt.Sprintf(text, args...)
	}
	return
}
