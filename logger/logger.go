package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"time"
)

type Logger struct {
	name      string
	usePath   string
	outputs   []Adapter
	callDepth int
}

func New(depth ...int) *Logger {
	dep := append(depth, 2)[0]
	l := &Logger{}
	//l.outputs = make(map[string]Adapter)
	l.callDepth = dep
	return l
}

func (this *Logger) Remove(i Adapter) {
	var outputs []Adapter
	for _, v := range this.outputs {
		if v != i {
			outputs = append(outputs, v)
		}
	}
	this.outputs = outputs
	i.Close()
}

//Adapter 增加输出接口
func (this *Logger) Adapter(i Adapter) error {
	if err := i.Init(); err != nil {
		return err
	}
	this.outputs = append(this.outputs, i)
	return nil
}

//SetLogPathTrim 设置日志起始路径
func (this *Logger) SetLogPathTrim(trimPath string) {
	this.usePath = filepath.ToSlash(trimPath)
}

func (this *Logger) writeToLoggers(msg *Message, level int) {
	for name, l := range this.outputs {
		if err := l.Write(msg, level); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "unable to WriteMsg to adapter:%v,error:%v\n", name, err)
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
	if len(v) > 0 {
		msg = fmt.Sprintf(msg, v...)
	}

	msgSt := new(Message)
	src := ""
	when := time.Now()
	if this.callDepth > 0 {
		_, file, lineno, ok := runtime.Caller(3)
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

func (this *Logger) Alert(format interface{}, args ...interface{}) {
	this.writeMsg(LevelAlert, format, args...)
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
