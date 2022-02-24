package main

import (
	"github.com/hwcer/cosgo/cosweb"
)

func main() {
	Server := cosweb.NewServer()
	Server.Register("/ping", ping)
	if err := Server.Run(":80"); err != nil {
		panic(err)
	}
}

func ping(context *cosweb.Context, f func()) error {
	return context.String("ok")
}
