package cosnet

import (
	"errors"
	"sync"
	"sync/atomic"
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
	NetTypeUdp                //UDP类型
	NetTypeWs                 //websocket
)

type Server interface {
	Start() error
	Close() error
	Stopped() bool
	Sockets() *Sockets
	Runtime() int64
	GetHandler() Handler
	GetMsgType() MsgType
	GetNetType() NetType
	SetMultiplex(bool)
	GetMultiplex() bool
}

func NewNetServer(msgTyp MsgType, handler Handler, netType NetType) *NetServer {
	s := &NetServer{
		wgp:     new(sync.WaitGroup),
		msgTyp:  msgTyp,
		netType: netType,
		handler: handler,
		sockets: new(Sockets),
	}
	s.startServerTicker()
	return s
}

type NetServer struct {
	wgp       *sync.WaitGroup
	stop      int32
	msgTyp    MsgType //消息类型
	netType   NetType
	address   string
	sockets   *Sockets //自增ID
	handler   Handler  //消息处理器
	multiplex bool     //是否使用协程来处理MESSAGE
	timestamp int64    //时间
}

func (s *NetServer) Close() error {
	if !atomic.CompareAndSwapInt32(&s.stop, 0, 1) {
		return errors.New("server stoping")
	}
	return nil
}

func (s *NetServer) Stopped() bool {
	return s.stop == 1
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

func (s *NetServer) SetMultiplex(multiplex bool) {
	s.multiplex = multiplex
}
func (s *NetServer) GetMultiplex() bool {
	return s.multiplex
}

//服务器运行时长,非精确时长
func (s *NetServer) Runtime() int64 {
	return s.timestamp
}

func (s *NetServer) startServerTicker() {
	Go(s.wgp, func() {
		t := time.Millisecond * time.Duration(Config.ServerInterval)
		ticker := time.NewTimer(t)
		defer ticker.Stop()
		for !s.Stopped() {
			select {
			case <-ticker.C:
				s.timestamp += Config.ServerInterval
				s.sockets.ticker()
				ticker.Reset(t)
			}
		}
	})
}
