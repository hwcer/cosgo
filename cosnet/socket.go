package cosnet

import (
	"cosgo/logger"
	"sync/atomic"
	"time"
)

//各种服务器(TCP,UDP,WS)也使用该接口
type Socket interface {
	Id() uint32

	LocalAddr() string
	RemoteAddr() string
	SetRealRemoteAddr(addr string)

	Close() bool
	Stoped() bool
	IsProxy() bool

	Write(m *Message) bool

	SetUser(interface{})
	GetUser() interface{}
}

func NewNetSocket(srv Server) *NetSocket {
	sockets := srv.Sockets()
	sock := &NetSocket{
		id:        sockets.Id(),
		cwrite:    make(chan *Message, Config.WriteChanSize),
		server:    srv,
		heartbeat: srv.Timestamp(),
	}
	sock.ticker = time.NewTicker(time.Millisecond * time.Duration(Config.ConnectHeartbeat))
	return sock
}

type NetSocket struct {
	id             uint32 //唯一标示
	stop           int32  //停止标记
	ticker         *time.Ticker
	cwrite         chan *Message //写入通道
	server         Server
	handler        Handler
	user           interface{} //玩家登陆后信息
	heartbeat      int64       //最后有效行为时间戳
	realRemoteAddr string      //当使用代理是，需要特殊设置客户端真实IP
}

func (s *NetSocket) Id() uint32 {
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
		close(s.cwrite)
	}
	s.ticker.Stop()
	s.server.Sockets().Del(s.Id())
	logger.Debug("socket Close Id:%d", s.id)
	return true
}

//判断连接是否关闭
func (s *NetSocket) Stoped() bool {
	if s.server.Stoped() {
		s.Close()
	}
	return s.stop > 0
}

func (s *NetSocket) IsProxy() bool {
	return s.realRemoteAddr != ""
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
	//if Config.AutoCompressLen > 0 && m.Head != nil && m.Head.Len >= Config.AutoCompressLen && !m.Head.HasFlag(MsgFlagCompress){
	//	m.Head.AddFlag(MsgFlagCompress)
	//	m.Data = GZipCompress(m.Data)
	//	m.Head.Len = uint32(len(m.Data))
	//}
	select {
	case s.cwrite <- m:
	default:
		logger.Warn("socket write channel full id:%v", s.id)
		s.Close() //通道已满，直接关闭
	}

	return true
}

func (s *NetSocket) timeout() bool {
	if int32(s.server.Timestamp()-s.heartbeat) > Config.ConnectTimeout {
		return true
	} else {
		return false
	}
}

func (s *NetSocket) processMsg(msgque Socket, msg *Message) bool {
	s.heartbeat = s.server.Timestamp()
	if s.server.GetMultiplex() {
		go s.processMsgTrue(msgque, msg)
	} else {
		return s.processMsgTrue(msgque, msg)
	}
	return true
}
func (s *NetSocket) processMsgTrue(sock Socket, msg *Message) bool {
	if msg.Head != nil && msg.Head.HasFlag(MsgFlagCompress) && msg.Data != nil {
		data, err := GZipUnCompress(msg.Data)
		if err != nil {
			logger.Error("uncompress failed socket:%v err:%v", sock.Id(), err)
			return false
		}
		msg.Data = data
		msg.Head.SubFlag(MsgFlagCompress)
		msg.Head.Len = uint32(len(msg.Data))
	}
	return s.handler.OnMessage(sock, msg)
}
