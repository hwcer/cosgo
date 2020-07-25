package antnet

import (
	"net"
	"reflect"
	"strings"
	"time"
)

var DefMsgQueTimeout int = 180 //MS

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

type ConnType int

const (
	ConnTypeListen ConnType = iota //服务器监听
	ConnTypeAccept                 //客户端CONN
)

//IMsgQue 相当于socket.io中的socket,客户端连接信息
//各种服务器(TCP,UDP,WS)也使用该接口，偷懒？
type IMsgQue interface {
	Id() uint32
	GetNetType() NetType
	GetMsgType() MsgType
	GetConnType() ConnType

	LocalAddr() string
	RemoteAddr() string
	SetRealRemoteAddr(addr string)

	Stop()
	IsStop() bool
	Available() bool
	IsProxy() bool

	Send(m *Message) (re bool)

	SetSendFast()
	SetTimeout(t int)

	GetTimeout() int
	Reconnect(t int) //重连间隔  最小1s，此函数仅能连接关闭时调用

	GetHandler() IMsgHandler

	SetUser(user interface{})
	GetUser() interface{}

	//服务器内部通讯时提升效率，比如战斗服发送消息到网关服，应该在连接建立时使用，cwriteCnt大于0表示重新设置cwrite缓存长度，内网一般发送较快，不用考虑
	SetMultiplex(multiplex bool, cwriteCnt int) bool
}

type msgQue struct {
	id uint32 //唯一标示

	cwrite  chan *Message //写入通道
	stop    int32         //停止标记
	msgTyp  MsgType       //消息类型
	connTyp ConnType      //通道类型

	handler  IMsgHandler //消息处理器
	timeout  int         //传输超时
	lastTick int64

	init      bool
	available bool
	sendFast  bool
	multiplex bool //是否使用协程来处理MESSAGE

	user           interface{} //玩家登陆后信息
	realRemoteAddr string      //当使用代理是，需要特殊设置客户端真实IP
}

func (r *msgQue) SetSendFast() {
	r.sendFast = true
}

func (r *msgQue) SetUser(user interface{}) {
	r.user = user
}

func (r *msgQue) Available() bool {
	return r.available
}

func (r *msgQue) GetUser() interface{} {
	return r.user
}

func (r *msgQue) GetHandler() IMsgHandler {
	return r.handler
}

func (r *msgQue) GetMsgType() MsgType {
	return r.msgTyp
}

func (r *msgQue) GetConnType() ConnType {
	return r.connTyp
}

func (r *msgQue) Id() uint32 {
	return r.id
}

func (r *msgQue) SetTimeout(t int) {
	if t >= 0 {
		r.timeout = t
	}
}

func (r *msgQue) isTimeout(tick *time.Timer) bool {
	left := int(Timestamp - r.lastTick)
	if left < r.timeout || r.timeout == 0 {
		if r.timeout == 0 {
			tick.Reset(time.Second * time.Duration(DefMsgQueTimeout))
		} else {
			tick.Reset(time.Second * time.Duration(r.timeout-left))
		}
		return false
	}
	Logger.Debug("msgque close because timeout id:%v wait:%v timeout:%v", r.id, left, r.timeout)
	return true
}

func (r *msgQue) GetTimeout() int {
	return r.timeout
}

func (r *msgQue) Reconnect(t int) {

}

func (r *msgQue) IsProxy() bool {
	return r.realRemoteAddr != ""
}

func (r *msgQue) SetRealRemoteAddr(addr string) {
	r.realRemoteAddr = addr
}

func (r *msgQue) SetMultiplex(multiplex bool, cwriteCnt int) bool {
	t := r.multiplex
	r.multiplex = multiplex
	if cwriteCnt > 0 {
		r.cwrite = make(chan *Message, cwriteCnt)
	}
	return t
}

func (r *msgQue) Send(m *Message) (re bool) {
	if m == nil {
		return
	}
	defer func() {
		if err := recover(); err != nil {
			re = false
		}
	}()
	if Config.AutoCompressLen > 0 && m.Head != nil && m.Head.Len >= Config.AutoCompressLen && (m.Head.Flags&FlagCompress) == 0 {
		m.Head.Flags |= FlagCompress
		m.Data = GZipCompress(m.Data)
		m.Head.Len = uint32(len(m.Data))
	}
	select {
	case r.cwrite <- m:
	default:
		Logger.Warn("msgque write channel full msgque:%v", r.id)
		//r.cwrite <- m
		r.baseStop() //通道已满，直接关闭
	}

	return true
}

func (r *msgQue) baseStop() {
	if r.cwrite != nil {
		close(r.cwrite)
	}
	msgqueMapSync.Lock()
	delete(msgqueMap, r.id)
	msgqueMapSync.Unlock()
	Logger.Debug("msgque close id:%d", r.id)
}
func (r *msgQue) processMsg(msgque IMsgQue, msg *Message) bool {
	if r.multiplex {
		Go(func() {
			r.processMsgTrue(msgque, msg)
		})
	} else {
		return r.processMsgTrue(msgque, msg)
	}
	return true
}
func (r *msgQue) processMsgTrue(msgque IMsgQue, msg *Message) bool {
	if msg.Head != nil && msg.Head.Flags&FlagCompress > 0 && msg.Data != nil {
		data, err := GZipUnCompress(msg.Data)
		if err != nil {
			Logger.Error("msgque uncompress failed msgque:%v act:%v len:%v err:%v", msgque.Id(),  msg.Head.Act, msg.Head.Len, err)
			return false
		}
		msg.Data = data
		msg.Head.Flags -= FlagCompress
		msg.Head.Len = uint32(len(msg.Data))
	}
	f := r.handler.GetHandlerFunc(msgque, msg)
	if f == nil {
		f = r.handler.OnProcessMsg
	}
	return f(msgque, msg)
}

type HandlerFunc func(msgque IMsgQue, msg *Message) bool

type IMsgHandler interface {
	OnNewMsgQue(msgque IMsgQue) bool                         //新的消息队列
	OnDelMsgQue(msgque IMsgQue)                              //消息队列关闭
	OnProcessMsg(msgque IMsgQue, msg *Message) bool          //默认的消息处理函数
	OnConnectComplete(msgque IMsgQue, ok bool) bool          //连接成功，client使用
	GetHandlerFunc(msgque IMsgQue, msg *Message) HandlerFunc //根据消息获得处理函数
}

type DefMsgHandler struct {
	msgMap  map[int]HandlerFunc
	typeMap map[reflect.Type]HandlerFunc
}

func (r *DefMsgHandler) OnNewMsgQue(msgque IMsgQue) bool                { return true }
func (r *DefMsgHandler) OnDelMsgQue(msgque IMsgQue)                     {}
func (r *DefMsgHandler) OnProcessMsg(msgque IMsgQue, msg *Message) bool { return true }
func (r *DefMsgHandler) OnConnectComplete(msgque IMsgQue, ok bool) bool { return true }
func (r *DefMsgHandler) GetHandlerFunc(msgque IMsgQue, msg *Message) HandlerFunc {
	return nil
}

func (r *DefMsgHandler) RegisterMsg(v interface{}, fun HandlerFunc) {
	msgType := reflect.TypeOf(v)
	if msgType == nil || msgType.Kind() != reflect.Ptr {
		Logger.Error("message pointer required")
		return
	}
	if r.typeMap == nil {
		r.typeMap = map[reflect.Type]HandlerFunc{}
	}
	r.typeMap[msgType] = fun
}

type EchoMsgHandler struct {
	DefMsgHandler
}

func (r *EchoMsgHandler) OnProcessMsg(msgque IMsgQue, msg *Message) bool {
	msgque.Send(msg)
	return true
}

func StartServer(addr string, typ MsgType, handler IMsgHandler) error {
	addrs := strings.Split(addr, "://")
	if addrs[0] == "tcp" || addrs[0] == "all" {
		listen, err := net.Listen("tcp", addrs[1])
		if err == nil {
			msgque := newTcpListen(listen, typ, handler, addr)
			Go(func() {
				Logger.Debug("process listen for tcp msgque:%d", msgque.id)
				msgque.listen()
				Logger.Debug("process listen end for tcp msgque:%d", msgque.id)
			})
		} else {
			Logger.Error("listen on %s failed, errstr:%s", addr, err)
			return err
		}
	}
	if addrs[0] == "udp" || addrs[0] == "all" {
		naddr, err := net.ResolveUDPAddr("udp", addrs[1])
		if err != nil {
			Logger.Error("listen on %s failed, errstr:%s", addr, err)
			return err
		}
		conn, err := net.ListenUDP("udp", naddr)
		if err == nil {
			msgque := newUdpListen(conn, typ, handler, addr)
			Go(func() {
				Logger.Debug("process listen for udp msgque:%d", msgque.id)
				msgque.listen()
				Logger.Debug("process listen end for udp msgque:%d", msgque.id)
			})
		} else {
			Logger.Error("listen on %s failed, errstr:%s", addr, err)
			return err
		}
	}
	if addrs[0] == "ws" || addrs[0] == "wss" {
		naddr := strings.SplitN(addrs[1], "/", 2)
		url := "/"
		if len(naddr) > 1 {
			url = "/" + naddr[1]
		}
		if addrs[0] == "wss" {
			Config.EnableWss = true
		}
		if typ != MsgTypeCmd {
			Logger.Error("ws type msgque noly support MsgTypeCmd now auto set to MsgTypeCmd")
		}
		msgque := newWsListen(naddr[0], url, MsgTypeCmd, handler)
		Go(func() {
			Logger.Debug("process listen for ws msgque:%d", msgque.id)
			msgque.listen()
			Logger.Debug("process listen end for ws msgque:%d", msgque.id)
		})
	}
	return nil
}
