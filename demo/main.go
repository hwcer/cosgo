package main

import (
	"cosgo/app"
	"cosgo/cosnet"
	"cosgo/cosweb"
	middleware "cosgo/cosweb/middleware"
	"cosgo/demo/handle"
	"cosgo/session"
	"github.com/spf13/pflag"
	"reflect"
)

func init() {
	pflag.String("tcp", "tcp://0.0.0.0:3100", "tcp address")
	pflag.String("http", "0.0.0.0:80", "http address")

	pflag.String("hwc", "", "test pflag")
	app.Config.SetDefault("pprof", "0.0.0.0:8080") //开启性能分析工具
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
	m.web.Options.SessionStorage = session.NewMemory(nil)
	m.web.Use(allMiddleware)
	m.web.Debug = false
	//使用Group并在每个Group上添加不同的中间件
	//g := cosweb.NewGroup()
	//g.Use(groupMiddleware)
	//g.SetCaller(caller)
	//g.Register(&handle.Remote{})
	//g.Route(m.web, "/")
	g2 := m.web.Group("/", &handle.Remote{})
	g2.Register(&handle.Admin{})
	g2.Use(groupMiddleware, adminMiddleware)
	g2.SetCaller(caller)

	access := middleware.NewAccessControlAllow()
	access.Origin = append(access.Origin, "*.baidu.com", "*", "163.com")
	access.Methods = append(access.Methods, "GET", "POST", "OPTIONS")
	g2.Use(access.Handle)

	//m.web.Proxy("/", "https://www.jd.com")
	m.web.Static("/static", "wwwroot")
	return
}

func allMiddleware(ctx *cosweb.Context, next func()) {
	//logger.Debug("do middleware")
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
func adminMiddleware(ctx *cosweb.Context, next func()) {
	path := ctx.Get(cosweb.RouteGroupPath)
	if path != "admin" {
		next()
		return
	}
	var level int = 2
	api := ctx.Get(cosweb.RouteGroupName)
	if api == "login" {
		level = 0
	}
	err := ctx.Session.Start(level)
	if err != nil {
		ctx.Error(err)
	} else {
		next()
	}
}

func caller(proto, method reflect.Value, c *cosweb.Context) error {
	k := c.Get(cosweb.RouteGroupPath)
	if k == "remote" {
		p := proto.Interface().(*handle.Remote)
		f := method.Interface().(func(*handle.Remote, *cosweb.Context) error)
		return f(p, c)
	} else if k == "admin" {
		p := proto.Interface().(*handle.Admin)
		f := method.Interface().(func(*handle.Admin, *cosweb.Context) error)
		return f(p, c)
	} else {
		return cosweb.ErrNotFound
	}
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
