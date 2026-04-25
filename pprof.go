package cosgo

import (
	"fmt"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/hwcer/cosgo/utils"
)

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package pprof serves via its HTTP server runtime profiling data
// in the format expected by the pprof visualization tool.
//
// The package is typically only imported for the side effect of
// registering its HTTP handlers.
// The handled paths all begin with /debug/pprof/.
//
// To use pprof, link this package into your program:
//
//	import _ "net/http/pprof"
//
// If your application is not already running an http server, you
// need to start one. Keys "net/http" and "log" to your imports and
// the following code to your main function:
//
//	go func() {
//		log.Println(http.ListenAndServe("localhost:6060", nil))
//	}()
//
// If you are not using DefaultServeMux, you will have to register handlers
// with the mux you are using.
//
// Then use the pprof tool to look at the heap profile:
//
//	go tool pprof http://localhost:6060/debug/pprof/heap
//
// Or to look at a 30-second CPU profile:
//
//	go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30
//
// Or to look at the goroutine blocking profile, after calling
// runtime.SetBlockProfileRate in your program:
//
//	go tool pprof http://localhost:6060/debug/pprof/block
//
// Or to look at the holders of contended mutexes, after calling
// runtime.SetMutexProfileFraction in your program:
//
//	go tool pprof http://localhost:6060/debug/pprof/mutex
//
// The package also exports a handler that serves execution trace data
// for the "go tool trace" command. To collect a 5-second execution trace:
//
//	wget -O trace.out http://localhost:6060/debug/pprof/trace?seconds=5
//	go tool trace trace.out
//
// To view all available profiles, open http://localhost:6060/debug/pprof/
// in your browser.
//
// For a study of the facility in action, visit
//
//	https://blog.golang.org/2011/06/profiling-go-programs.html
var pprofServer *http.Server

// pprofStart 在 config.pprof 配置了监听地址时启动独立的 pprof HTTP 服务。
// 刻意不使用 `import _ "net/http/pprof"` 的副作用注册,避免 pprof handler
// 污染 http.DefaultServeMux(否则业务 server 若复用 DefaultServeMux 会意外暴露
// /debug/pprof 端点)。改为手动注册到专用 mux,接入面仅限本服务地址。
func pprofStart() error {
	addr := Config.GetString("pprof")
	if addr == "" {
		return nil
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	pprofServer = &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}
	_ = utils.Timeout(time.Second, func() error {
		return pprofServer.ListenAndServe()
	})
	fmt.Printf("pprof server start:%v\n", addr)
	return nil
}

func pprofClose() error {
	if pprofServer == nil {
		return nil
	}
	fmt.Printf("pprof server close\n")
	return pprofServer.Close()
}
