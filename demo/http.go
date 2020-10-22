package main

import (
	"cosgo/express"
	"cosgo/logger"
	"fmt"
	"sync"
)

func NewHttpMod(name string) *httpMod {
	return &httpMod{name: name}
}

type httpMod struct {
	app  *express.Engine
	name string
}

func (this *httpMod) ID() string {
	return this.name
}

func (this *httpMod) Load() error {
	this.app = express.New("")
	//this.app.Use(middleware1, middleware2)

	this.app.Debug = true
	this.app.GET("/favicon.ioc", favicon)
	this.app.Any("/s/:api/*", hello)
	this.app.Group([]string{express.HttpMethodAny}, "/nsp/", &remote{})

	//代理
	//this.app.Proxy("*", "http://redis.cn/")
	return nil
}

func (this *httpMod) Start(wgp *sync.WaitGroup) (err error) {
	wgp.Add(1)
	this.app.Start()
	return
}

func (this *httpMod) Close(wgp *sync.WaitGroup) error {
	this.app.Close()
	wgp.Done()
	return nil
}

// handler
func hello(c *express.Context) error {
	logger.Debug("Hello, World")
	return c.String(fmt.Sprintf("Hello, World 1!  %v\n", c.Params()))
}

func favicon(c *express.Context) error {
	return c.Status(404).End()
}
func middleware1(c *express.Context, next func()) {
	logger.Debug("middleware1")
	next()
}

func middleware2(c *express.Context, next func()) {
	logger.Debug("middleware2")
	next()
}

func middleware3(c *express.Context, next func()) {
	logger.Debug("middleware3")
	next()
}
