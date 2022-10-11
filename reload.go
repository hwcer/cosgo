package cosgo

import (
	"fmt"
	"github.com/hwcer/logger"
)

func Reload() {
	fmt.Printf("SIGHUP reload Config\n")
	for _, m := range modules {
		if err := m.Reload(); err != nil {
			logger.Warn("[%v]reload error:%v", m.ID(), err)
		}
	}
}
