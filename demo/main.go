package main

import (
	"demo/module"
	"github.com/hwcer/cosgo/app"
)

func init() {
	app.Config.SetDefault("pprof", "0.0.0.0:8080") //开启性能分析工具
}

func main() {
	//app.SetGCSummaryTime(10)
	app.Use(module.NewTcpModule("TCP"))
	app.Use(module.NewWebModule("HTTP"))
	app.Start()
}
