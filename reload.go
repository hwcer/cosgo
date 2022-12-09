package cosgo

import (
	"github.com/hwcer/cosgo/logger"
)

func Reload() {
	console := logger.NewConsoleAdapter()
	_ = logger.SetAdapter(console)
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
		logger.DelAdapter(console)
	}()
	logger.Info("SIGHUP reload Config")
	for _, m := range modules {
		if err := m.Reload(); err != nil {
			logger.Warn("[%v]reload error:%v", m.ID(), err)
		}
	}
}
