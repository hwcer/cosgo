package cosnet

import (
	"cosgo/logger"
	"sync/atomic"
)

//各种服务器(TCP,UDP,WS)也使用该接口
type Socket interface {
	Id() uint64
	LocalAddr() string
	RemoteAddr() string
	SetRealRemoteAddr(addr string)
	Close() bool
	Stopped() bool
	IsProxy() bool
	Write(m *Message) bool
	SetUser(interface{})
	GetUser() interface{}
	KeepAlive()
	init(uint64)
	timeout()
}

func NewSocket(handler Handler) (sock *NetSocket) {
	sock = &NetSocket{
		cwrite:    make(chan *Message, Config.WriteChanSize),
		handler:   handler,
		timestamp: timestamp,
	}
	return
}

type NetSocket struct {
	id             uint64        //唯一标示
	user           interface{}   //玩家登陆后信息
	stop           int32         //停止标记
	cwrite         chan *Message //写入通道
	handler        Handler       //消息处理器
	timestamp      int           //最后有效行为时间戳
	realRemoteAddr string        //当使用代理是，需要特殊设置客户端真实IP
}

func (s *NetSocket) init(id uint64) {
	s.id = id
}

func (s *NetSocket) timeout() {
	if timestamp-s.timestamp >= Config.ConnectTimeout {
		s.Close()
	}
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

//
func (s *NetSocket) Close() bool {
	if !atomic.CompareAndSwapInt32(&s.stop, 0, 1) {
		return false
	}
	if s.cwrite != nil {
		close(s.cwrite) //关闭cwrite强制write协程取消堵塞快速响应关闭操作
	}
	s.handler.OnDisconnect(s)
	logger.Debug("Socket Close Id:%d", s.id)
	return true
}

//判断连接是否关闭
func (s *NetSocket) Stopped() bool {
	if s.stop == 0 && SCC.Stopped() {
		s.Close()
	}
	return s.stop > 0
}

func (s *NetSocket) IsProxy() bool {
	return s.realRemoteAddr != ""
}

func (s *NetSocket) KeepAlive() {
	s.timestamp = timestamp
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
		s.Close() //通道已满，直接关闭
	}

	return true
}

func (s *NetSocket) processMsg(sock Socket, msg *Message) bool {
	s.KeepAlive()
	if msg.Head != nil && msg.Head.Flags.Has(MsgFlagCompress) && msg.Data != nil {
		data, err := GZipUnCompress(msg.Data)
		if err != nil {
			logger.Error("uncompress failed socket:%v err:%v", sock.Id(), err)
			return false
		}
		msg.Data = data
		msg.Head.Flags.Del(MsgFlagCompress)
		msg.Head.Size = int32(len(msg.Data))
	}
	return s.handler.OnMessage(sock, msg)
}
