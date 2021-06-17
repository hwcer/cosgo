package module

import (
	"demo/handle"
	"github.com/hwcer/cosgo/app"
	"github.com/hwcer/cosgo/cosweb"
	"github.com/hwcer/cosgo/cosweb/middleware"
	"github.com/hwcer/cosgo/cosweb/render"
	"github.com/spf13/pflag"
	"reflect"
)

func init() {
	pflag.String("http", "0.0.0.0:80", "http address")
}

func NewWebModule(id string) *webModule {
	return &webModule{
		DefModule: app.DefModule{Id: id},
	}
}

type webModule struct {
	app.DefModule
	web *cosweb.Server
}

func (m *webModule) Init() (err error) {
	//cosweb.Options.Debug = true
	http := app.Config.GetString("http")

	m.web = cosweb.NewServer(http)
	m.web.Render = cosweb.NewRender(&render.Options{
		TemplatesGlob: app.GetDir() + "/wwwroot/*.html",
		Debug:         cosweb.Options.Debug,
		Delims:        []string{"<%", "%>"},
	})
	m.web.Use(allMiddleware1, allMiddleware2)
	//m.web.Debug = true
	//使用Group并在每个Group上添加不同的中间件
	//g := cosweb.NewGroup()
	//g.Use(groupMiddleware)
	//g.SetCaller(caller)
	//g.Register(&handle.Remote{})
	//g.Route(m.web, "/")
	g2 := m.web.Group("/", &handle.Admin{})
	g2.Register(&handle.Admin{})
	g2.Use(groupMiddleware, adminMiddleware)
	g2.SetCaller(caller)

	access := middleware.NewAccessControlAllow()
	access.Origin = append(access.Origin, "*.baidu.com", "*", "163.com")
	access.Methods = append(access.Methods, "GET", "POST", "OPTIONS")
	g2.Use(access.Handle)

	m.web.GET("/s/*", func(c *cosweb.Context, next func()) error {
		return c.String("test")
	})

	//m.web.Proxy("/", "https://www.jd.com")
	m.web.Static("/", app.GetDir()+"/wwwroot")
	return
}

func (m *webModule) Start() error {
	return m.web.Start()
}

func (m *webModule) Close() error {
	return m.web.Close()
}

func allMiddleware1(ctx *cosweb.Context, next func()) {
	next()
	//logger.Debug("do middleware1")
	header := ctx.Response.Header()
	header.Set("X-TEST-HEADER", "TEST")
	//ctx.String("STRING:middleware return\n")
}
func allMiddleware2(ctx *cosweb.Context, next func()) {
	next()
	//logger.Debug("do middleware2")
	header := ctx.Response.Header()
	header.Set("X-TEST-HEADER", "TEST")
	//ctx.String("STRING:middleware return\n")
}
func groupMiddleware(ctx *cosweb.Context, next func()) {
	next()
	//logger.Debug("do group middleware")

	//ctx.String("group middleware return\n")
}
func adminMiddleware(ctx *cosweb.Context, next func()) {
	//logger.Debug("do admin middleware")
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

func caller(proto, method reflect.Value, c *cosweb.Context) (err error) {
	k := c.Get(cosweb.RouteGroupPath)
	if k == "remote" {
		p := proto.Interface().(*handle.Remote)
		f := method.Interface().(func(*handle.Remote, *cosweb.Context) error)
		err = f(p, c)
	} else if k == "admin" {
		p := proto.Interface().(*handle.Admin)
		f := method.Interface().(func(*handle.Admin, *cosweb.Context) error)
		err = f(p, c)
	}
	return
}
