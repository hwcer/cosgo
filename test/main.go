package main

import (
	"context"
	"cosgo/app"
	"cosgo/logger"
	"sync"
)
func init()  {
	logger.Debug("test main init")
}

type test struct {
	wgp *sync.WaitGroup
	name string
}

func (this *test)ID()string  {
	return this.name
}

func (this *test)Ready()error  {
	return nil
}

func (this *test)Start(cx context.Context, wgp *sync.WaitGroup) error {
	this.wgp = wgp
	this.wgp.Add(1)
	return nil
}

func (this *test)Stop()error  {
	this.wgp.Done()
	return nil
}



func main() {
	app.SetMain(func() {
		logger.Debug("程序启动啦")
	})

	app.Start(&test{name:"testMod"})
}