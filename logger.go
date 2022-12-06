package cosgo

import (
	logger "github.com/hwcer/logger"
	"strings"
)

func init() {
	logger.DefaultLogger.SetCallDepth(0)
	if v, ok := logger.DefaultAdapter.(*logger.ConsoleAdapter); ok {
		v.Format = loggerMessageFormat
	}
	logger.DefaultLogger.SetLogPathTrim(workDir)
}

func loggerMessageFormat(msg *logger.Message) string {
	b := strings.Builder{}
	b.WriteString(msg.Time.Format(Options.DataTimeFormat))
	b.WriteString(" [")
	b.WriteString(msg.Level)
	b.WriteString("] ")
	b.WriteString(msg.Content)
	return b.String()
}

// removeDefaultLogger 移除默认控制台日志输出
func removeConsoleLogger() {
	logger.DefaultLogger.Remove(logger.DefaultAdapter)
}
