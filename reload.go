package cosgo

import (
	"github.com/hwcer/logger"
)

func Reload() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
		logger.Trace("Reload Config Finish\n")
	}()
	logger.Trace("Start reload Config\n")
	emit(EventTypReload)
	for _, m := range modules {
		if reload, ok := m.(ModuleReload); ok {
			if err := reload.Reload(); err != nil {
				logger.Trace("[%v]reload error:%v\n", m.Id(), err)
			}
		}
	}
}
