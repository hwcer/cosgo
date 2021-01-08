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

	Send(m *Message) bool

	SetUser(interface{})
	GetUser() interface{}
}

func NewNetSocket(srv Server) *NetSocket {
	sockets := srv.Sockets()
	return &NetSocket{
		id:        sockets.Id(),
		cwrite:    make(chan *Message, Config.WriteChanSize),
		heartbeat: srv.Timestamp(),
		server:    srv,
	}
}

type NetSocket struct {
	id      uint32 //唯一标示
	stop    int32  //停止标记
	ticker  *time.Ticker
	cwrite  chan *Message //写入通道
	server  Server
	handler MsgHandler

	user interface{} //玩家登陆后信息

	heartbeat int64 //最后有效行为时间戳

	realRemoteAddr string //当使用代理是，需要特殊设置客户端真实IP
}

func (r *NetSocket) Id() uint32 {
	return r.id
}

func (r *NetSocket) tickerStop() {
	r.ticker.Stop()
}
func (r *NetSocket) tickerStart() {
	r.ticker = time.NewTicker(time.Millisecond * time.Duration(Config.ConnectHeartbeat))
}

func (r *NetSocket) SetUser(user interface{}) {
	r.user = user
}

func (r *NetSocket) GetUser() interface{} {
	return r.user
}

//判断连接是否关闭
func (r *NetSocket) Stoped() bool {
	if r.server.Stoped() {
		r.Close()
	}
	return r.stop > 0
}

func (r *NetSocket) IsProxy() bool {
	return r.realRemoteAddr != ""
}

func (r *NetSocket) SetRealRemoteAddr(addr string) {
	r.realRemoteAddr = addr
}

func (r *NetSocket) Send(m *Message) (re bool) {
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
	case r.cwrite <- m:
	default:
		logger.Warn("socket write channel full id:%v", r.id)
		r.Close() //通道已满，直接关闭
	}

	return true
}

//
func (r *NetSocket) Close() bool {
	if !atomic.CompareAndSwapInt32(&r.stop, 0, 1) {
		return false
	}
	if r.cwrite != nil {
		close(r.cwrite)
	}
	r.server.Sockets().Del(r.Id())
	logger.Debug("socket Close Id:%d", r.id)
	return true
}

func (r *NetSocket) timeout() bool {
	if int32(r.server.Timestamp()-r.heartbeat) > Config.ConnectTimeout {
		return true
	} else {
		return false
	}
}

func (r *NetSocket) processMsg(msgque Socket, msg *Message) bool {
	r.heartbeat = r.server.Timestamp()
	if r.server.GetMultiplex() {
		go r.processMsgTrue(msgque, msg)
	} else {
		return r.processMsgTrue(msgque, msg)
	}
	return true
}
func (r *NetSocket) processMsgTrue(msgque Socket, msg *Message) bool {
	if msg.Head != nil && msg.Head.HasFlag(MsgFlagCompress) && msg.Data != nil {
		data, err := GZipUnCompress(msg.Data)
		if err != nil {
			logger.Error("socket uncompress failed msgque:%v act:%v len:%v err:%v", msgque.Id(), msg.Head.Proto, msg.Head.Len, err)
			return false
		}
		msg.Data = data
		msg.Head.SubFlag(MsgFlagCompress)
		msg.Head.Len = uint32(len(msg.Data))
	}
	f := r.handler.GetHandlerFunc(msgque, msg)
	if f == nil {
		f = r.handler.OnProcessMsg
	}
	return f(msgque, msg)
}
