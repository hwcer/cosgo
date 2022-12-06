package cosgo

import (
	"github.com/hwcer/logger"
)

func Reload() {
	_ = logger.DefaultLogger.Adapter(logger.DefaultAdapter)
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
		logger.DefaultLogger.Remove(logger.DefaultAdapter)
	}()
	logger.Info("SIGHUP reload Config")
	for _, m := range modules {
		if err := m.Reload(); err != nil {
			logger.Warn("[%v]reload error:%v", m.ID(), err)
		}
	}
}
