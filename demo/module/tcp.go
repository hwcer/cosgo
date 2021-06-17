package module

import (
	"fmt"
	"github.com/hwcer/cosgo/app"
	"github.com/hwcer/cosgo/cosnet"
	"github.com/hwcer/cosgo/cosnet/message"
	"github.com/spf13/pflag"
)

func init() {
	message.SetAttachField("test", 4)
	pflag.String("tcp", "tcp://0.0.0.0:3100", "tcp address")
}
func NewTcpModule(id string) *tcp {
	return &tcp{
		DefModule: app.DefModule{Id: id},
	}
}

type tcp struct {
	app.DefModule
	srv cosnet.Server
}

func (m *tcp) Init() (err error) {
	addr := app.Config.GetString("tcp")
	m.srv = cosnet.NewServer(addr, &TcpHandler{})
	return
}

func (m *tcp) Start() error {
	return m.srv.Start()
}
func (m *tcp) Close() error {
	return m.srv.Close()
}

type TcpHandler struct {
}

func (this *TcpHandler) Message(sock cosnet.Socket, msg *message.Message) {
	sock.Write(msg)
	fmt.Printf("Message head:%+v,attach:%v,data:%v\n", msg.Head, msg.Attach, msg.Data)
}
