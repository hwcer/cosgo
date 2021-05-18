package cosnet

import (
	"context"
	"github.com/hwcer/cosgo/utils"
	"time"
)

type MsgType int

const (
	MsgTypeMsg MsgType = iota //消息基于确定的消息头
	MsgTypeCmd                //消息没有消息头，以\n分割
)

type NetType int

const (
	NetTypeTcp NetType = iota //TCP类型
	NetTypeUdp                //UDP类型dw
	NetTypeWss                //websocket
)

type Server interface {
	On(EventsType, EventsFunc)
	Emit(EventsType, Socket)
	Start() error
	Close() error
	Handler() Handler
	Context() context.Context
	Sockets() *Sockets
	GetMsgType() MsgType
	GetNetType() NetType
}

func NewNetServer(address string, handler Handler, msgTyp MsgType, netType NetType) *NetServer {
	s := &NetServer{
		SCC:     utils.NewSCC(nil),
		Events:  NewEvents(),
		msgTyp:  msgTyp,
		netType: netType,
		address: address,
		handler: handler,
		sockets: NewSockets(handler, 1024),
	}
	s.On(EventsTypeConnect, func(sock Socket) {
		sock.(*TcpSocket).id = s.sockets.Add(sock)
	})
	s.On(EventsTypeDisconnect, func(sock Socket) {
		s.sockets.Del(sock.Id())
	})
	s.CGO(s.sockets.Start)
	return s
}

type NetServer struct {
	*utils.SCC
	*Events
	msgTyp  MsgType //消息类型
	netType NetType
	address string
	handler Handler
	sockets *Sockets
}

func (s *NetServer) Close() error {
	if !s.SCC.Close() {
		return nil
	}
	return s.SCC.Wait(time.Second * 10)
}

func (s *NetServer) Handler() Handler {
	return s.handler
}

func (s *NetServer) Sockets() *Sockets {
	return s.sockets
}

func (s *NetServer) GetMsgType() MsgType {
	return s.msgTyp
}

func (s *NetServer) GetNetType() NetType {
	return s.netType
}
