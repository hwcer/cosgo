package main

import (
	"cosgo/demo/handle"
	"cosgo/express"
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
	this.app.Group("/", &handle.Remote{})
	this.app.RESTful("/", &handle.Restful{})

	this.app.Static("/web/", "web")
	//所有不认识的路由都转向给redis.cn
	this.app.Proxy("/", "http://redis.cn")
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
