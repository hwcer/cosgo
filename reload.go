package cosgo

import (
	"github.com/hwcer/logger"
)

// Reload 模块重载接口，实现此接口的模块支持运行时重载
type Reload interface {
	// Reload 模块重载，在应用收到重载信号时调用
	// 此阶段主要进行模块的配置重新加载、状态刷新等操作
	Reload() error
}

func reload() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
		logger.Trace("Reload Config Finish\n")
	}()
	logger.Trace("Start reload Config\n")
	emit(EventTypReload)
	for _, m := range modules {
		if r, ok := m.(Reload); ok {
			if err := r.Reload(); err != nil {
				logger.Trace("[%v]reload error:%v\n", m.Id(), err)
			}
		}
	}
}
