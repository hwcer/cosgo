package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type Logger struct {
	mutex     sync.Mutex
	level     Level
	usePath   []string
	outputs   map[string]Output
	callDepth int
}

func New(depth ...int) *Logger {
	dep := append(depth, 2)[0]
	l := &Logger{}
	l.level = LevelError
	l.outputs = map[string]Output{}
	l.callDepth = dep
	return l
}

func (this *Logger) Write(msg *Message) {
	defer func() {
		_ = recover()
	}()
	if msg.Level < this.level {
		return
	}
	if msg.Time.IsZero() {
		msg.Time = time.Now()
	}
	if this.callDepth > 0 && msg.Path == "" {
		if _, file, lineno, ok := runtime.Caller(this.callDepth); ok {
			file = this.TrimPath(file)
			msg.Path = strings.Replace(fmt.Sprintf("%s:%d", file, lineno), "%2e", ".", -1)
		}
	}
	if msg.Level >= LevelError && msg.Stack == "" {
		msg.Stack = string(debug.Stack())
	}
	for name, output := range this.outputs {
		if err := output.Write(msg); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "unable to WriteMsg to adapter:%v,error:%v\n", name, err)
		}
	}
}

func (this *Logger) Sprint(level Level, format any, args ...any) {
	this.Write(&Message{Content: Sprintf(format, args...), Level: level})
}

func (this *Logger) Fatal(format any, args ...any) {
	this.Sprint(LevelFATAL, format, args...)
	os.Exit(1)
}

func (this *Logger) Panic(format any, args ...any) {
	this.Sprint(LevelPanic, format, args...)
	panic(Sprintf(format, args...))
}

// Error Log ERROR level message.
func (this *Logger) Error(format interface{}, v ...interface{}) {
	this.Sprint(LevelError, format, v...)
}
func (this *Logger) Alert(format interface{}, args ...interface{}) {
	this.Sprint(LevelAlert, format, args...)
}

// Debug Log DEBUG level message.
func (this *Logger) Debug(format interface{}, v ...interface{}) {
	this.Sprint(LevelDebug, format, v...)
}

// Trace Log TRAC level message.
func (this *Logger) Trace(format interface{}, v ...interface{}) {
	this.Sprint(LevelTrace, format, v...)
}

// SetLevel 设置日志输出等级
func (this *Logger) SetLevel(level Level) {
	this.level = level
}

// SetPathTrim 设置日志起始路径
func (this *Logger) SetPathTrim(trimPath ...string) {
	for _, p := range trimPath {
		this.usePath = append(this.usePath, filepath.ToSlash(p))
	}
}

func (this *Logger) SetCallDepth(depth int) {
	this.callDepth = depth
}

func (this *Logger) TrimPath(s string) (r string) {
	if len(this.usePath) == 0 {
		return s
	}
	r = s
	for _, p := range this.usePath {
		arr := strings.SplitN(r, p, 2)
		if len(arr) == 2 {
			arr[0] = p
			r = strings.Join(arr, "")
		}
	}
	return
}
