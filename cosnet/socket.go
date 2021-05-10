package cosnet

import (
	"context"
	"cosgo/logger"
	"cosgo/utils"
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

func NewSocket(ctx context.Context, handler Handler) (sock *NetSocket) {
	sock = &NetSocket{
		cwrite:  make(chan *Message, Config.WriteChanSize),
		handler: handler,
		timeout: Config.SocketTimeout,
	}
	sock.ctx, sock.cancel = context.WithCancel(ctx)
	return
}

type NetSocket struct {
	id     uint64 //唯一标示
	ctx    context.Context
	cancel context.CancelFunc

	user interface{} //玩家登陆后信息
	//stop    int32         //停止标记,0:正常,1-掉线（等待短线重连）,2-销毁 无法再短线重连
	cwrite  chan *Message //写入通道
	handler Handler       //消息处理器

	timeout        int    //超时
	heartbeat      int    //  heartbeat >=timeout 时被标记为超时
	realRemoteAddr string //当使用代理是，需要特殊设置客户端真实IP
}

func (s *NetSocket) Id() uint64 {
	return s.id
}

func (s *NetSocket) Done() bool {
	return utils.Done(s.ctx)
}

func (s *NetSocket) SetUser(user interface{}) {
	s.user = user
}

func (s *NetSocket) GetUser() interface{} {
	return s.user
}

//关闭
func (s *NetSocket) Close() bool {
	s.cancel()
	//if s.cwrite != nil {
	//	close(s.cwrite) //关闭cwrite强制write协程取消堵塞快速响应关闭操作
	//}
	logger.Debug("Socket Close Id:%d", s.id)
	return true
}

//销毁
//func (s *NetSocket) Destroy() bool {
//	if !atomic.CompareAndSwapInt32(&s.stop, 1, 2) {
//		return false
//	}
//	return true
//}

func (s *NetSocket) IsProxy() bool {
	return s.realRemoteAddr != ""
}

//每一次Heartbeat() heartbeat计数加1
func (s *NetSocket) Heartbeat() {
	s.heartbeat += 1
	if s.heartbeat >= s.timeout {
		s.Close()
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
	return s.handler.Message(s.ctx, sock, msg)
}
