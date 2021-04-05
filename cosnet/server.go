package cosnet

import (
	"sync"
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
	NetTypeWs                 //websocket
)

type Server interface {
	Start() error
	Close() error
	Stopped() bool
	Sockets() *Sockets
	GetHandler() Handler
	GetMsgType() MsgType
	GetNetType() NetType
}

func NewNetServer(msgTyp MsgType, handler Handler, netType NetType) *NetServer {
	s := &NetServer{
		wgp:     new(sync.WaitGroup),
		msgTyp:  msgTyp,
		netType: netType,
		handler: handler,
		sockets: NewSockets(1024),
	}
	return s
}

type NetServer struct {
	wgp     *sync.WaitGroup
	msgTyp  MsgType //消息类型
	netType NetType
	address string
	handler Handler //消息处理器
	sockets *Sockets
}

func (s *NetServer) Close() error {
	return s.sockets.Close()
}

func (s *NetServer) Stopped() bool {
	return s.sockets.Stopped()
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

func (s *NetServer) GetHandler() Handler {
	return s.handler
}
