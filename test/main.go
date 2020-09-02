package main

import (
	"context"
	"cosgo/app"
	"cosgo/logger"
	"cosgo/network"
	"encoding/json"
	"github.com/spf13/pflag"
	"net/url"
	"sync"
	"time"
)

func init()  {
	pflag.String("hwc","","test pflag")
	app.Flag.SetDefault("proAddr","0.0.0.0:8080")  //开启性能分析工具
	logger.Debug("test main init")
}

type test struct {
	name string
}

func (this *test)ID()string  {
	return this.name
}

func (this *test)Init()error  {

	return nil
}

func (this *test)Start(ctx context.Context, wgp *sync.WaitGroup) error {

	hwc:=app.Flag.GetString("hwc")
	logger.Debug("FLAG HWC:%v",hwc)

	s:= "http://127.0.0.1:8001/api/王权富贵?x=y&"
	p,_ := url.Parse(s)

	query := p.Query()
	query.Set("a","%")
	query.Set("b","h=东方月初")
	p.RawQuery = query.Encode()


	js,_:=json.Marshal(p)
	logger.Debug("URL OBJ:%+v",string(js))

	logger.Debug("URL STR:%v",p.String())
	logger.Debug("URL PATH:%v",p.EscapedPath())
	logger.Debug("=========================很严肃的分界线=======================")

	t:= time.Now()
	logger.Debug("时间:%v",t.Format("20060102"))
	logger.Debug("=========================启动个网关试试=======================")
	network.StartServer("tcp://0.0.0.0:8201",network.MsgTypeMsg,&network.DefMsgHandler{})

	return nil
}

func (this *test)Stop()error  {
	network.Stop()
	return nil
}



func main() {
	app.SetMain(func() {
		logger.Debug("程序启动啦")
	})

	app.Start(&test{name:"testMod"})
}