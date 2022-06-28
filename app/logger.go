package app

import (
	logger2 "github.com/hwcer/cosgo/logger"
)

var (
	loggerFileAdapter     *logger2.FileAdapter
	loggerConsoleAdapter  *logger2.ConsoleAdapter
	loggerFileAdapterName string
)

func init() {
	loggerFileAdapterName = "cosgoLoggerFileAdapterName"
	loggerConsoleAdapter, _ = logger2.Default.Get(logger2.DefaultAdapterName).(*logger2.ConsoleAdapter)
	if loggerConsoleAdapter != nil {
		loggerConsoleAdapter.Options.Format = loggerMessageFormat
	}
	logger2.SetLogPathTrim(workDir)
}

//setLogger 将日志从控制台转移到日志文件
func setLogger() {
	if loggerConsoleAdapter != nil {
		loggerConsoleAdapter.Options.Format = nil
	}
	//设置日志
	if loggerFileAdapter != nil {
		if err := logger2.Adapter(loggerFileAdapterName, loggerFileAdapter); err == nil {
			logger2.Remove(logger2.DefaultAdapterName)
		}
	}
}
