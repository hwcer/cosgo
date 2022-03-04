package cosrpc

import (
	"fmt"
	"github.com/hwcer/cosgo/library/logger"
	"github.com/smallnest/rpcx/log"
)

func init() {
	l := &dummyLogger{}
	log.SetLogger(l)
}

type dummyLogger struct {
}

func (l *dummyLogger) Debug(v ...interface{}) {
	logger.Debug(fmt.Sprint(v))
}

func (l *dummyLogger) Debugf(format string, v ...interface{}) {
	logger.Debug(format, v...)
}

func (l *dummyLogger) Info(v ...interface{}) {
	logger.Info(fmt.Sprint(v))
}

func (l *dummyLogger) Infof(format string, v ...interface{}) {
	logger.Info(format, v...)
}

func (l *dummyLogger) Warn(v ...interface{}) {
	logger.Warn(fmt.Sprint(v))
}

func (l *dummyLogger) Warnf(format string, v ...interface{}) {
	logger.Warn(format, v...)
}

func (l *dummyLogger) Error(v ...interface{}) {
	logger.Error(fmt.Sprint(v))
}

func (l *dummyLogger) Errorf(format string, v ...interface{}) {
	logger.Error(format, v...)
}

func (l *dummyLogger) Fatal(v ...interface{}) {
	logger.Fatal(fmt.Sprint(v))
}

func (l *dummyLogger) Fatalf(format string, v ...interface{}) {
	logger.Fatal(format, v...)
}

func (l *dummyLogger) Panic(v ...interface{}) {
	logger.Panic(fmt.Sprint(v))
}

func (l *dummyLogger) Panicf(format string, v ...interface{}) {
	logger.Panic(format, v...)
}
