package main

import (
	"cosgo/express"
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
	this.express.Any("/s/:api", hello)
	this.express.Proxy("/proxy/*", "http://127.0.0.1:7902")
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
	return c.String(fmt.Sprintf("Hello, World 1!  %v\n", c.Params()))
}
