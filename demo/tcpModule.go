package main

import (
	"context"
	"cosgo/app"
	"cosgo/cosnet"
	"cosgo/logger"
	"github.com/spf13/pflag"
)

func init() {
	pflag.String("tcp", "tcp://0.0.0.0:3100", "tcp address")
}
func NewTcpModule(id string) *tcpModule {
	return &tcpModule{
		DefModule: app.DefModule{Id: id},
	}
}

type tcpModule struct {
	app.DefModule
	srv cosnet.Server
}

func (m *tcpModule) Init() (err error) {
	addr := app.Config.GetString("tcp")
	m.srv = cosnet.NewServer(addr, NewTcpHandler())
	return
}

func (m *tcpModule) Start() error {
	return m.srv.Start()
}
func (m *tcpModule) Close() error {
	return m.srv.Close()
}

func NewTcpHandler() *TcpHandler {
	return &TcpHandler{
		HandlerDefault: cosnet.NewHandlerDefault(),
	}
}

type TcpHandler struct {
	*cosnet.HandlerDefault
}

func (this *TcpHandler) Message(ctx context.Context, sock cosnet.Socket, msg *cosnet.Message) bool {
	logger.Debug("OnMessage:%+v", msg)
	return true
}
