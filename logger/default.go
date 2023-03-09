package logger

import "path/filepath"

const DefTimeFormat = "2006-01-02 15:04:05-0700" // 日志输出默认格式
const defaultConsoleAdapterName = "_defaultConsoleAdapter"

var defaultLogger *Logger

func init() {
	defaultLogger = New(3)
	console := NewConsoleAdapter()
	console.name = defaultConsoleAdapterName
	_ = defaultLogger.SetAdapter(console)
}

func Alert(f interface{}, v ...interface{}) {
	defaultLogger.Alert(f, v...)
}

func Fatal(f interface{}, v ...interface{}) {
	defaultLogger.Fatal(f, v...)
}

func Error(f interface{}, v ...interface{}) {
	defaultLogger.Error(f, v...)
}

func Warn(f interface{}, v ...interface{}) {
	defaultLogger.Warn(f, v...)
}

func Info(f interface{}, v ...interface{}) {
	defaultLogger.Info(f, v...)
}

func Debug(f interface{}, v ...interface{}) {
	defaultLogger.Debug(f, v...)
}

func Trace(f interface{}, v ...interface{}) {
	defaultLogger.Trace(f, v...)
}

func SetAdapter(i Adapter) error {
	return defaultLogger.SetAdapter(i)
}
func DelAdapter(i interface{}) {
	defaultLogger.DelAdapter(i)
}

// SetLogLevel 设置日志输出等级
func SetLogLevel(level Level) {
	defaultLogger.level = level
}

// SetLogPathTrim 设置日志起始路径
func SetLogPathTrim(trimPath string) {
	defaultLogger.usePath = filepath.ToSlash(trimPath)
}
func SetCallDepth(depth int) {
	defaultLogger.callDepth = depth
}

func GetDefaultLogger() *Logger {
	return defaultLogger
}

func GetDefaultAdapter() *ConsoleAdapter {
	for _, v := range defaultLogger.outputs {
		if v.Name() == defaultConsoleAdapterName {
			return v.(*ConsoleAdapter)
		}
	}
	return nil
}

func DelDefaultAdapter() {
	defaultLogger.DelAdapter(defaultConsoleAdapterName)
}
