package app

import (
	"github.com/hwcer/cosgo/library/logger"
)

var (
	loggerFileAdapter     *logger.FileAdapter
	loggerConsoleAdapter  *logger.ConsoleAdapter
	loggerFileAdapterName string
)

func init() {
	loggerFileAdapterName = "cosgoLoggerFileAdapterName"
	loggerConsoleAdapter, _ = logger.Default.Get(logger.DefaultAdapterName).(*logger.ConsoleAdapter)
	if loggerConsoleAdapter != nil {
		loggerConsoleAdapter.Options.Format = loggerMessageFormat
	}
	logger.SetLogPathTrim(workDir)
}

//setLogger 将日志从控制台转移到日志文件
func setLogger() {
	if loggerConsoleAdapter != nil {
		loggerConsoleAdapter.Options.Format = nil
	}
	//设置日志
	if loggerFileAdapter != nil {
		if err := logger.Adapter(loggerFileAdapterName, loggerFileAdapter); err == nil {
			logger.Remove(logger.DefaultAdapterName)
		}
	}
}
