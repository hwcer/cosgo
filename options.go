package cosgo

import (
	"github.com/hwcer/logger"
)

var Options = &struct {
	Banner  func()
	Process func() bool //设置启动进程，返回false时不会继续向下执行
	//DataTimeFormat string
}{
	Banner: defaultBanner,
	//DataTimeFormat: "2006-01-02 15:04:05-0700",
}

// SetBanner 设置默认启动界面，启动完所有MOD后执行
func SetBanner(f func()) {
	Options.Banner = f
}

func SetProcess(f func() bool) {
	Options.Process = f
}

func defaultBanner() {
	logger.Trace("congratulations on successful startup\n")
}
