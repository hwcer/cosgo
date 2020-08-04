package server

import (
	"context"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func NewHttpServer() *httpServer {
	return &httpServer{
		Engine: gin.New(),
		Router:NewRouter(),
	}
}

//外网服务器配置
type HttpServerOption struct {
	Tls    bool   `json:"tls"`    // 使用tls
	Addr   string `json:"addr"`   // 网关地址
	TlsPem string `json:"tlspem"` // 证书 pem
	TlsKey string `json:"tlskey"` // 证书 key
}

type httpServer struct {
	Router *Router
	Engine *gin.Engine
	Server *http.Server
	Options *HttpServerOption
}



func (s *httpServer) Start(opts *HttpServerOption) (err error) {
	s.Options = opts
	s.Server = &http.Server{
		Addr:           s.Options.Addr,
		Handler:        s.Engine,
		ReadTimeout:    7 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1MB
	}

	//s.Engine.POST("/*",s.Router.Handle)
	if s.Options.Tls {
		pemFile := s.Options.TlsPem
		if pemFile == "" {
			panic("httpServer Options[TlsPem] empty")
		}
		keyFile := s.Options.TlsKey
		if keyFile == "" {
			panic("httpServer Options[TlsKey] empty")
		}
		err = TimeOut(time.Second,func() error {
			e := s.Server.ListenAndServeTLS(pemFile, keyFile)
			if e != nil && e != http.ErrServerClosed {
				return e
			}
			return nil
		})
	} else {
		err = TimeOut(time.Second,func() error {
			e := s.Server.ListenAndServe()
			if e != nil && e != http.ErrServerClosed {
				return e
			}
			return nil
		})
	}
	return
}

func (s *httpServer) Stop(ctx context.Context) (err error) {
	return s.Server.Shutdown(ctx)
}



func (s *httpServer) Register(handle interface{}) *nsp{
	return s.Router.Register(handle)
}
