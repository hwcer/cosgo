package cosgo

import (
	"fmt"
	"github.com/hwcer/cosgo/logger"
	"strings"
)

func init() {
	logger.SetCallDepth(0)
	logger.SetLogPathTrim(workDir)
	SetLoggerFormat(loggerFormatDefault)
}

func SetLoggerFormat(f func(*logger.Message) string) {
	if v := logger.GetDefaultAdapter(); v != nil {
		v.Format = f
	}
}

func loggerFormatDefault(msg *logger.Message) string {
	b := strings.Builder{}
	//b.WriteString(msg.Time.Format(logger.DefTimeFormat))
	//b.WriteString(" [")
	//b.WriteString(msg.Level.String())
	//b.WriteString("] ")
	b.WriteString(msg.Content)
	return b.String()
}

var Console = console{}

type console struct {
	silent bool
}

// Close 关闭控制台输出
func (this *console) Close() {
	this.silent = true
}

func (this *console) Fatal(f interface{}, v ...interface{}) {
	if !this.silent {
		logger.Fatal(f, v...)
	}
}

func (this *console) Error(f interface{}, v ...interface{}) {
	if !this.silent {
		logger.Error(f, v...)
	}
}

func (this *console) Warn(f interface{}, v ...interface{}) {
	if !this.silent {
		logger.Warn(f, v...)
	}
}

func (this *console) Info(f interface{}, v ...interface{}) {
	if !this.silent {
		logger.Info(f, v...)
	}
}

func (this *console) Debug(f interface{}, v ...interface{}) {
	if !this.silent {
		logger.Debug(f, v...)
	}
}

func (this *console) Printf(format string, a ...any) {
	if !this.silent {
		fmt.Printf(format, a...)
	}
}
