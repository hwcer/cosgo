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
	name    string
	express *express.Engine
}

func (this *httpMod) ID() string {
	return this.name
}

func (this *httpMod) Load() error {
	this.express = express.New("")
	this.express.Use(middleware1, middleware2)

	this.express.Debug = true
	this.express.Any("/s/*", hello, middleware3)
	this.express.Any("/s/*", hello, middleware3)
	//代理
	//this.express.Proxy("*", "http://redis.cn/")
	return nil
}

func (this *httpMod) Start(wgp *sync.WaitGroup) (err error) {
	wgp.Add(1)
	this.express.Start()
	return
}

func (this *httpMod) Close(wgp *sync.WaitGroup) error {
	this.express.Close()
	wgp.Done()
	return nil
}

// handler
func hello(c *express.Context) error {
	logger.Debug("Hello, World")
	return c.String(fmt.Sprintf("Hello, World 1!  %v\n", c.Params()))
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
