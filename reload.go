package cosgo

import (
	"github.com/hwcer/cosgo/logger"
)

func Reload() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
		logger.Trace("Reload Config Finish")
	}()
	logger.Trace("Start reload Config")
	for _, m := range modules {
		if err := m.Reload(); err != nil {
			logger.Alert("[%v]reload error:%v", m.ID(), err)
		}
	}
}
