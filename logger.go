package cosgo

import (
	logger "github.com/hwcer/logger"
	"strings"
)

var (
	loggerFileAdapter *logger.FileAdapter
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

// setLogger 将日志从控制台转移到日志文件
func setLogger() {
	if loggerFileAdapter != nil {
		//loggerFileAdapter.Format = loggerMessageFormat
		if err := logger.DefaultLogger.Adapter(loggerFileAdapter); err != nil {
			logger.Error(err)
		} else {
			logger.DefaultLogger.Remove(logger.DefaultAdapter)
		}
	}

}
