package main

import (
	"cosgo/express"
	"cosgo/logger"
	"net/http"
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
	return nil
}

func (this *httpMod) Start(wgp *sync.WaitGroup) (err error) {
	wgp.Add(1)
	this.express.GET("/", hello)
	return this.express.Start()
}

func (this *httpMod) Close(wgp *sync.WaitGroup) error {
	this.express.Close()
	wgp.Done()
	return nil
}

func middleware1(c *express.Context, next func()) {
	logger.Debug("middleware1")
	next()
}
func middleware2(c *express.Context, next func()) {
	logger.Debug("middleware2")
	c.HTML(200, "hahah")
	next()
}

// Handler
func hello(c *express.Context) error {
	return c.String(http.StatusOK, "Hello, World!")
}
