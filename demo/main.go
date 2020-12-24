package main

import (
	"cosgo/app"
	"cosgo/cosnet"
	"cosgo/demo/handle"
	"cosgo/express"
	"github.com/spf13/pflag"
)

func init() {
	pflag.String("tcp", "tcp://0.0.0.0:3100", "tcp address")
	pflag.String("http", "0.0.0.0:80", "http address")

	pflag.String("hwc", "", "test pflag")
	app.Config.SetDefault("proAddr", "0.0.0.0:8080") //开启性能分析工具
}

func main() {
	tcp := cosnet.New("tpc srv", "", nil)
	tcp.Flag = "tcp"

	http := express.New("HTTPSRV", "", &handle.Remote{})
	http.Flag = "http"
	app.Start()
}
