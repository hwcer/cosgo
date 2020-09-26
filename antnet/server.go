package antnet

import (
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
	GetHandler() IMsgHandler
	GetMsgType() MsgType
	GetNetType() NetType
	SetMultiplex(bool)
	GetMultiplex() bool
}

type defServer struct {
	msgTyp    MsgType //消息类型
	netType   NetType
	handler   IMsgHandler //消息处理器
	multiplex bool        //是否使用协程来处理MESSAGE
}

func (r *defServer) GetMsgType() MsgType {
	return r.msgTyp
}

func (r *defServer) GetNetType() NetType {
	return r.netType
}

func (r *defServer) GetHandler() IMsgHandler {
	return r.handler
}

func (r *defServer) SetMultiplex(multiplex bool) {
	r.multiplex = multiplex
}
func (r *defServer) GetMultiplex() bool {
	return r.multiplex
}

//各种服务器(TCP,UDP,WS)也使用该接口
type Socket interface {
	Id() uint32

	LocalAddr() string
	RemoteAddr() string
	SetRealRemoteAddr(addr string)

	Stop()
	IsStop() bool
	IsProxy() bool

	Send(m *Message) (re bool)

	SetUser(user interface{})
	GetUser() interface{}
}

type defSocket struct {
	id      uint32        //唯一标示
	stop    int32         //停止标记
	cwrite  chan *Message //写入通道
	server  Server
	handler IMsgHandler

	user interface{} //玩家登陆后信息

	heartbeat int64 //最后有效行为时间戳

	realRemoteAddr string //当使用代理是，需要特殊设置客户端真实IP
}

func (r *defSocket) Id() uint32 {
	return r.id
}

func (r *defSocket) SetUser(user interface{}) {
	r.user = user
}

func (r *defSocket) GetUser() interface{} {
	return r.user
}

//判断连接是否关闭
func (r *defSocket) IsStop() bool {
	if r.stop == 0 && IsStop() {
		r.Stop()
	}
	return r.stop == 1
}

func (r *defSocket) IsProxy() bool {
	return r.realRemoteAddr != ""
}

func (r *defSocket) SetRealRemoteAddr(addr string) {
	r.realRemoteAddr = addr
}

func (r *defSocket) Send(m *Message) (re bool) {
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
		Logger.Warn("msgque write channel full msgque:%v", r.id)
		r.Stop() //通道已满，直接关闭
	}

	return true
}

//
func (r *defSocket) Stop() {
	if !atomic.CompareAndSwapInt32(&r.stop, 0, 1) {
		return
	}
	if r.cwrite != nil {
		close(r.cwrite)
	}
	msgqueMapSync.Lock()
	delete(msgqueMap, r.id)
	msgqueMapSync.Unlock()
	Logger.Debug("msgque close id:%d", r.id)
}

func (r *defSocket) isTimeout(tick *time.Timer) bool {
	left := int32(timestamp - r.heartbeat)
	if left < Config.ConnectTimeout {
		tick.Reset(time.Millisecond * time.Duration(Config.ConnectHeartbeat))
		return false
	}
	Logger.Debug("msgque close because timeout id:%v wait:%v timeout:%v", r.id, left, Config.ConnectTimeout)
	return true
}

func (r *defSocket) processMsg(msgque Socket, msg *Message) bool {
	if r.server.GetMultiplex() {
		go r.processMsgTrue(msgque, msg)
	} else {
		return r.processMsgTrue(msgque, msg)
	}
	return true
}
func (r *defSocket) processMsgTrue(msgque Socket, msg *Message) bool {
	if msg.Head != nil && msg.Head.HasFlag(MsgFlagCompress) && msg.Data != nil {
		data, err := GZipUnCompress(msg.Data)
		if err != nil {
			Logger.Error("msgque uncompress failed msgque:%v act:%v len:%v err:%v", msgque.Id(), msg.Head.Proto, msg.Head.Len, err)
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
