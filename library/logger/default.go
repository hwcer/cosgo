package logger

// 默认日志输出
const DefTimeFormat = "2006-01-02 15:04:05 -0700" // 日志输出默认格式
const DefaultAdapterName string = "default"

var Default *Logger

func init() {
	Default = New(3)
	opts := NewConsoleOptions()
	console, _ := NewConsoleAdapter(opts)
	Default.Adapter(DefaultAdapterName, console)
}

func SetLogPathTrim(trimPath string) {
	Default.SetLogPathTrim(trimPath)
}

// Panic logs a message at emergency level and panic.
func Panic(f interface{}, v ...interface{}) {
	Default.Panic(f, v...)
}

// Fatal logs a message at emergency level and exit.
func Fatal(f interface{}, v ...interface{}) {
	Default.Fatal(f, v...)
}

// Error logs a message at error level.
func Error(f interface{}, v ...interface{}) {
	Default.Error(f, v...)
}

// Warn logs a message at warning level.
func Warn(f interface{}, v ...interface{}) {
	Default.Warn(f, v...)
}

// Info logs a message at info level.
func Info(f interface{}, v ...interface{}) {
	Default.Info(f, v...)
}

// Notice logs a message at debug level.
func Debug(f interface{}, v ...interface{}) {
	Default.Debug(f, v...)
}

// Trace logs a message at trace level.
func Trace(f interface{}, v ...interface{}) {
	Default.Trace(f, v...)
}

func Adapter(name string, i adapter) error {
	return Default.Adapter(name, i)
}
func Remove(name string) {
	Default.Remove(name)
}
