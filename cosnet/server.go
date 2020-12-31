package cosnet

import (
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
	Close() bool
	Stoped() bool
	Sockets() *sockets
	Timestamp() int64

	GetHandler() MsgHandler
	GetMsgType() MsgType
	GetNetType() NetType
	SetMultiplex(bool)
	GetMultiplex() bool
}

func NewNetServer(msgTyp MsgType, handler MsgHandler, netType NetType) *NetServer {
	s := &NetServer{
		wgp:     &sync.WaitGroup{},
		msgTyp:  msgTyp,
		netType: netType,
		handler: handler,
		cancel:  make(chan struct{}),
	}
	go s.goTick()
	return s
}

type NetServer struct {
	wgp     *sync.WaitGroup
	stop    int32
	cancel  chan struct{}
	ticker  *time.Ticker
	sockets *sockets //自增ID

	msgTyp    MsgType //消息类型
	netType   NetType
	address   string
	handler   MsgHandler //消息处理器
	timestamp int64      //时间
	multiplex bool       //是否使用协程来处理MESSAGE
}

func (s *NetServer) Close() bool {
	if !atomic.CompareAndSwapInt32(&s.stop, 0, 1) {
		return false
	}
	close(s.cancel)
	return true
}
func (s *NetServer) Stoped() bool {
	return s.stop > 0
}
func (s *NetServer) Sockets() *sockets {
	return s.sockets
}
func (s *NetServer) Timestamp() int64 {
	return s.timestamp
}

func (s *NetServer) goTick() {
	if s.ticker != nil {
		return
	}
	s.ticker = time.NewTicker(time.Millisecond * time.Duration(Config.ServerInterval))
	defer s.ticker.Stop()
	for {
		select {
		case <-s.ticker.C:
			s.timestamp += Config.ServerInterval
		case <-s.cancel:
			return
		}
	}
}

func (r *NetServer) GetMsgType() MsgType {
	return r.msgTyp
}

func (r *NetServer) GetNetType() NetType {
	return r.netType
}

func (r *NetServer) GetHandler() MsgHandler {
	return r.handler
}

func (r *NetServer) SetMultiplex(multiplex bool) {
	r.multiplex = multiplex
}
func (r *NetServer) GetMultiplex() bool {
	return r.multiplex
}
