package main

import (
	"cosgo/app"
)

func init() {
	app.Config.SetDefault("pprof", "0.0.0.0:8080") //开启性能分析工具
}

func main() {
	app.SetGCSummaryTime(10)
	app.Use(NewTcpModule("TCP"))
	//app.Use(NewWebModule("HTTP"))
	app.Start()
}
