package main

import (
	"cosgo/app"
	"cosgo/cosnet"
	"cosgo/cosweb"
	"cosgo/demo/handle"
	"cosgo/logger"
	"github.com/spf13/pflag"
	"reflect"
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
	http := app.Config.GetString("http")
	m.web = cosweb.NewServer(http)
	m.web.Use(middleware)
	m.web.Debug = true
	//使用Group并在每个Group上添加不同的中间件
	//g := cosweb.NewGroup()
	//g.Use(groupMiddleware)
	//g.SetCaller(caller)
	//g.Register(&handle.Remote{})
	//g.Route(m.web, "/")
	g2 := m.web.Group("/", &handle.Remote{})
	g2.Use(groupMiddleware)
	g2.SetCaller(caller)

	m.web.Proxy("/", "https://www.jd.com")
	m.web.Static("/static", "wwwroot")
	return
}

func middleware(ctx *cosweb.Context, next func()) {
	logger.Debug("do middleware")
	header := ctx.Response.Header()
	header.Set("X-TEST-HEADER", "TEST")
	//ctx.String("STRING:middleware return\n")
	next()
}

func groupMiddleware(ctx *cosweb.Context, next func()) {
	//logger.Debug("do group middleware")

	//ctx.String("group middleware return\n")
	next()
}

func caller(proto, method reflect.Value, c *cosweb.Context) error {
	p := proto.Interface().(*handle.Remote)
	f := method.Interface().(func(*handle.Remote, *cosweb.Context) error)
	return f(p, c)
}

func (m *module) Start() error {
	m.web.Start()
	return nil
}

func (m *module) Close() error {
	return m.web.Close()
}

func main() {

	app.Use(&module{Id: "test"})

	app.Start()
}
