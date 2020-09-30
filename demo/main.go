package main

import (
	"cosgo/app"
	"github.com/spf13/pflag"
)

func init() {

	pflag.String("hwc", "", "test pflag")
	app.Flag.SetDefault("proAddr", "0.0.0.0:8080") //开启性能分析工具
}

func main() {
	app.Use(NewTcpMod("TCPSRV"))
	app.Use(NewHttpMod("HTTPSRV"))
	app.Start()
}
