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
	level     Level
	usePath   string
	outputs   []Adapter
	callDepth int
}

func New(depth ...int) *Logger {
	dep := append(depth, 2)[0]
	l := &Logger{}
	l.level = LevelError
	l.callDepth = dep
	return l
}

// DelAdapter 删除Adapter,i =Adapter | Adapter name
func (this *Logger) DelAdapter(i interface{}) {
	var v Adapter
	var k string
	switch t := i.(type) {
	case string:
		k = t
	default:
		if v, _ = i.(Adapter); v != nil {
			k = v.Name()
		}
	}
	if k == "" {
		return
	}
	var outputs []Adapter
	for _, a := range this.outputs {
		if a.Name() == k {
			v = a
		} else {
			outputs = append(outputs, a)
		}
	}
	this.outputs = outputs
	if v != nil {
		v.Close()
	}
}

// Adapter 增加输出接口
func (this *Logger) SetAdapter(i Adapter) error {
	if err := i.Init(); err != nil {
		return err
	}
	//检查是否已经存在
	for _, v := range this.outputs {
		if v.Name() == i.Name() {
			return nil
		}
	}
	this.outputs = append(this.outputs, i)
	return nil
}

// SetLogLevel 设置日志输出等级
func (this *Logger) SetLogLevel(level Level) {
	this.level = level
}

// SetLogPathTrim 设置日志起始路径
func (this *Logger) SetLogPathTrim(trimPath string) {
	this.usePath = filepath.ToSlash(trimPath)
}

func (this *Logger) writeToLoggers(msg *Message) {
	for name, l := range this.outputs {
		if err := l.Write(msg); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "unable to WriteMsg to adapter:%v,error:%v\n", name, err)
		}
	}
}

func (this *Logger) writeMsg(level Level, err interface{}, v ...interface{}) {
	if level < this.level {
		return
	}
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
	msgSt.Level = level
	msgSt.Path = src
	msgSt.Content = msg
	msgSt.Time = when
	this.writeToLoggers(msgSt)
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
