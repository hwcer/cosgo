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

type sockets struct {
	mu        sync.Mutex
	index     uint32
	socketMap map[uint32]Socket
}

func (s *sockets) Id() uint32 {
	return atomic.AddUint32(&s.index, 1)
}

func (s *sockets) Add(sock Socket) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.socketMap[sock.Id()] = sock
}
func (s *sockets) Get(id uint32) Socket {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.socketMap[id]
}

func (s *sockets) Del(id ...uint32) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, k := range id {
		delete(s.socketMap, k)
	}
}

type Server interface {
	Start() error
	Close() error
	Stoped() bool
	Sockets() *sockets
	Timestamp() int64
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
		sockets: new(sockets),
	}
	s.starTimestampTicker()
	return s
}

type NetServer struct {
	wgp       *sync.WaitGroup
	stop      int32
	ticker    *time.Ticker
	timestamp int64 //时间

	msgTyp    MsgType //消息类型
	netType   NetType
	address   string
	sockets   *sockets //自增ID
	handler   Handler  //消息处理器
	multiplex bool     //是否使用协程来处理MESSAGE
}

func (s *NetServer) Close() error {
	if !atomic.CompareAndSwapInt32(&s.stop, 0, 1) {
		return errors.New("server stoping")
	}
	return nil
}

func (s *NetServer) Stoped() bool {
	return s.stop == 1
}

func (s *NetServer) Sockets() *sockets {
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
func (s *NetServer) Timestamp() int64 {
	return s.timestamp
}

func (s *NetServer) starTimestampTicker() {
	Go(s.wgp, func() {
		s.ticker = time.NewTicker(time.Millisecond * time.Duration(Config.ServerInterval))
		defer s.ticker.Stop()
		for !s.Stoped() {
			select {
			case <-s.ticker.C:
				s.timestamp += Config.ServerInterval
			}
		}
	})
}
