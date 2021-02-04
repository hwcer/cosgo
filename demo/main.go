package main

import (
	"cosgo/apps"
	"cosgo/cosnet"
	"cosgo/cosweb"
	"cosgo/demo/handle"
	"github.com/spf13/pflag"
	"reflect"
	"sync"
)

func init() {
	pflag.String("tcp", "tcp://0.0.0.0:3100", "tcp address")
	pflag.String("http", "0.0.0.0:80", "http address")

	pflag.String("hwc", "", "test pflag")
	//apps.Config.SetDefault("proAddr", "0.0.0.0:8080") //开启性能分析工具
}

type module struct {
	Id  string
	web *cosweb.Server
	srv cosnet.Server
}

func (this *module) ID() string {
	return this.Id
}

func (m *module) Init() (err error) {
	//addr := apps.Config.GetString("tcp")
	//m.srv, err = cosnet.NewServer(addr, nil)
	http := apps.Config.GetString("http")
	m.web = cosweb.NewServer(http)
	m.web.Debug = true
	g := m.web.Group("/", nil)
	m.web.Proxy("/", "https://www.jianshu.com")
	m.web.Static("/static", "wwwroot")

	//封装caller 避免使用reflect.Value.Call()方法
	g.Register(&handle.Remote{}, caller)
	return
}

func caller(proto, method reflect.Value, c *cosweb.Context) error {
	p := proto.Interface().(*handle.Remote)
	f := method.Interface().(func(*handle.Remote, *cosweb.Context) error)
	return f(p, c)
}

func (m *module) Start(wg *sync.WaitGroup) error {
	wg.Add(1)
	m.web.Start()
	return nil
}

func (m *module) Close(wg *sync.WaitGroup) error {
	defer wg.Done()
	return m.srv.Close()
}

func main() {

	apps.Use(&module{Id: "test"})

	apps.Start()
}
