package cosnet

import (
	"context"
	"fmt"
	"github.com/hwcer/cosgo/logger"
)

//各种服务器(TCP,UDP,WS)也使用该接口
type Socket interface {
	Id() uint64
	LocalAddr() string
	RemoteAddr() string
	SetRealRemoteAddr(addr string)
	Close() bool
	IsProxy() bool
	Write(m *Message) bool
	SetUser(interface{})
	GetUser() interface{}

	Heartbeat()
	KeepAlive()
}

func NewSocket(s Server) (sock *NetSocket) {
	sock = &NetSocket{
		cwrite:  make(chan *Message, Config.WriteChanSize),
		server:  s,
		timeout: Config.SocketTimeout,
	}
	sock.ctx, sock.cancel = context.WithCancel(s.Context())
	return
}

type NetSocket struct {
	id             uint64 //唯一标示
	ctx            context.Context
	cancel         context.CancelFunc
	server         Server
	user           interface{}   //玩家登陆后信息
	cwrite         chan *Message //写入通道
	timeout        int           //超时
	heartbeat      int           //  heartbeat >=timeout 时被标记为超时
	realRemoteAddr string        //当使用代理是，需要特殊设置客户端真实IP
}

func (s *NetSocket) Id() uint64 {
	return s.id
}

func (s *NetSocket) SetUser(user interface{}) {
	s.user = user
}

func (s *NetSocket) GetUser() interface{} {
	return s.user
}

//关闭
func (s *NetSocket) Close() bool {
	if s.Stopped() {
		return false
	}
	s.cancel()
	logger.Debug("Socket Close Id:%d", s.id)
	return true
}

func (s *NetSocket) Stopped() bool {
	select {
	case <-s.ctx.Done():
		return true
	default:
		return false
	}
}

func (s *NetSocket) IsProxy() bool {
	return s.realRemoteAddr != ""
}

//每一次Heartbeat() heartbeat计数加1
func (s *NetSocket) Heartbeat() {
	s.heartbeat += 1
	fmt.Printf("Heartbeat,id:%v,v:%v\n", s.id, s.heartbeat)
	if s.heartbeat >= s.timeout {
		s.cancel()
	}
}

//任何行为都清空heartbeat
func (s *NetSocket) KeepAlive() {
	s.heartbeat = 0
}

func (s *NetSocket) LocalAddr() string {
	return ""
}
func (s *NetSocket) RemoteAddr() string {
	return ""
}

func (s *NetSocket) SetRealRemoteAddr(addr string) {
	s.realRemoteAddr = addr
}

func (s *NetSocket) Write(m *Message) (re bool) {
	if m == nil {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			re = false
		}
	}()
	if Config.AutoCompressSize > 0 && m.Head != nil && m.Head.Size >= Config.AutoCompressSize && !m.Head.Flags.Has(MsgFlagCompress) {
		m.Head.Flags.Add(MsgFlagCompress)
		m.Data = GZipCompress(m.Data)
		m.Head.Size = int32(len(m.Data))
	}
	select {
	case s.cwrite <- m:
	default:
		logger.Warn("socket write channel full id:%v", s.id)
		s.cancel() //通道已满，直接关闭
	}

	return true
}

func (s *NetSocket) processMsg(sock Socket, msg *Message) {
	s.KeepAlive()
	if msg.Head != nil && msg.Head.Flags.Has(MsgFlagCompress) && msg.Data != nil {
		data, err := GZipUnCompress(msg.Data)
		if err != nil {
			s.cancel()
			logger.Error("uncompress failed socket:%v err:%v", sock.Id(), err)
			return
		}
		msg.Data = data
		msg.Head.Flags.Del(MsgFlagCompress)
		msg.Head.Size = int32(len(msg.Data))
	}
	handler := s.server.Handler()
	handler.Message(sock, msg)
}
