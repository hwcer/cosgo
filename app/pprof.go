package app

import (
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"time"
)

var pprofServer *http.Server

func pprofStart() error {
	pprof := Config.GetString("pprof")
	if pprof == "" {
		return nil
	}
	pprofServer = new(http.Server)
	pprofServer.Addr = pprof
	pprofServer.Handler = http.DefaultServeMux
	Timeout(time.Second, func() error {
		return pprofServer.ListenAndServe()
	})
	fmt.Printf("pprof server start:%v\n", pprof)
	return nil
}

func pprofClose() error {
	if pprofServer == nil {
		return nil
	}
	fmt.Printf("pprof server close\n")
	return pprofServer.Close()
}
