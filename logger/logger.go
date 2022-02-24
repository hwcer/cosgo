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

type Message struct {
	Path    string
	Time    time.Time
	Level   string
	Stack   string
	Content string
}

func (this *Message) String() string {
	return this.Time.Format(DefTimeFormat) + " [" + this.Level + "] " + "[" + this.Path + "] " + this.Content
}

type Logger struct {
	lock      sync.Mutex
	name      string
	usePath   string
	outputs   map[string]adapter
	callDepth int
}

func New(depth ...int) *Logger {
	dep := append(depth, 2)[0]
	l := &Logger{}
	l.outputs = make(map[string]adapter)
	l.callDepth = dep
	return l
}

func (this *Logger) Get(name string) adapter {
	return this.outputs[name]
}

func (this *Logger) Remove(name string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	delete(this.outputs, name)
}

//Adapter 增加输出接口
func (this *Logger) Adapter(name string, i adapter) error {
	if _, ok := this.outputs[name]; ok {
		return fmt.Errorf("logger adapter name exist:%v", name)
	}
	this.lock.Lock()
	defer this.lock.Unlock()
	outputs := make(map[string]adapter)
	outputs[name] = i
	for k, v := range this.outputs {
		outputs[k] = v
	}
	this.outputs = outputs
	return nil
}

// 设置日志起始路径
func (this *Logger) SetLogPathTrim(trimPath string) {
	this.usePath = filepath.ToSlash(trimPath)
}

func (this *Logger) writeToLoggers(msg *Message, level int) {
	for name, l := range this.outputs {
		if err := l.Write(msg, level); err != nil {
			fmt.Fprintf(os.Stderr, "unable to WriteMsg to adapter:%v,error:%v\n", name, err)
		}
	}
}

func (this *Logger) writeMsg(level int, err interface{}, v ...interface{}) string {
	var msg string
	switch err.(type) {
	case string:
		msg = err.(string)
	case error:
		msg = err.(error).Error()
	default:
		msg = fmt.Sprintf("%v", err)
	}

	msgSt := new(Message)
	src := ""
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}
	when := time.Now()
	if this.callDepth > 0 {
		_, file, lineno, ok := runtime.Caller(this.callDepth)
		if ok {
			if this.usePath != "" {
				file = stringTrim(file, this.usePath)
			}
			src = strings.Replace(fmt.Sprintf("%s:%d", file, lineno), "%2e", ".", -1)
		}
	}
	if level >= LevelError {
		msgSt.Stack = string(debug.Stack())
	}
	msgSt.Level = levelPrefix[level]
	msgSt.Path = src
	msgSt.Content = msg
	msgSt.Time = when
	this.writeToLoggers(msgSt, level)

	return msg
}

func (this *Logger) Fatal(format interface{}, args ...interface{}) {
	this.writeMsg(LevelFATAL, format, args...)
	os.Exit(1)
}

func (this *Logger) Panic(format interface{}, args ...interface{}) {
	msg := this.writeMsg(LevelPANIC, format, args...)
	panic(msg)
}

// Error Log ERROR level message.
func (this *Logger) Error(format interface{}, v ...interface{}) {
	this.writeMsg(LevelError, format, v...)
}

// Warn Log WARNING level message.
func (this *Logger) Warn(format interface{}, v ...interface{}) {
	this.writeMsg(LevelWarn, format, v...)
}

// Info Log INFO level message.
func (this *Logger) Info(format interface{}, v ...interface{}) {
	this.writeMsg(LevelInfo, format, v...)
}

// Debug Log DEBUG level message.
func (this *Logger) Debug(format interface{}, v ...interface{}) {
	this.writeMsg(LevelDebug, format, v...)
}

// Trace Log TRAC level message.
func (this *Logger) Trace(format interface{}, v ...interface{}) {
	this.writeMsg(LevelTrace, format, v...)
}

func (this *Logger) Close() {
	for _, l := range this.outputs {
		l.Close()
	}
	this.outputs = nil
}

func (this *Logger) SetCallDepth(depth int) {
	this.callDepth = depth
}

func stringTrim(s string, cut string) string {
	ss := strings.SplitN(s, cut, 2)
	if 1 == len(ss) {
		return ss[0]
	}
	return ss[1]
}
