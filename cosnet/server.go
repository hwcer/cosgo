package cosnet

import (
	"cosgo/utils"
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
	Start() error
	Close() error
	Sockets() *Sockets
	GetMsgType() MsgType
	GetNetType() NetType
}

func NewNetServer(address string, handler Handler, msgTyp MsgType, netType NetType) *NetServer {
	s := &NetServer{
		SCC:     utils.NewSCC(nil),
		msgTyp:  msgTyp,
		netType: netType,
		address: address,
		handler: handler,
		sockets: NewSockets(handler, 1024),
	}
	s.handler.On(HandlerEventTypeConnect, func(sock Socket) bool {
		sock.(*TcpSocket).id = s.sockets.Add(sock)
		return true
	})
	s.handler.On(HandlerEventTypeDisconnect, func(sock Socket) bool {
		return s.sockets.Del(sock.Id())
	})
	s.CGO(s.sockets.Start)
	return s
}

type NetServer struct {
	*utils.SCC
	msgTyp  MsgType //消息类型
	netType NetType
	address string
	handler Handler
	sockets *Sockets
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
