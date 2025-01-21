package cosgo

import (
	"fmt"
	"github.com/hwcer/cosgo/logger"
)

func Reload() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error(err)
		}
		fmt.Printf("Reload Config Finish\n")
	}()
	fmt.Printf("Start reload Config\n")
	for _, m := range modules {
		if err := m.Reload(); err != nil {
			fmt.Printf("[%v]reload error:%v\n", m.Id(), err)
		}
	}
}
