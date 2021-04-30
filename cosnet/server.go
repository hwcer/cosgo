package cosnet

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
	start() error
	Sockets() *Sockets
	GetMsgType() MsgType
	GetNetType() NetType
}

func NewNetServer(address string, handler Handler, msgTyp MsgType, netType NetType) *NetServer {
	s := &NetServer{
		msgTyp:  msgTyp,
		netType: netType,
		address: address,
		handler: handler,
		sockets: NewSockets(handler, 1024),
	}
	return s
}

type NetServer struct {
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
