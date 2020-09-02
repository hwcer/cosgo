package network

import (
	"sync"
	"sync/atomic"
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

type IMsgServer interface {
	loop() bool
	Stop()
	SetTimeout(int)
	GetTimeout() int
	GetHandler() IMsgHandler
	GetMsgType() MsgType
	GetNetType() NetType
	SetMultiplex(bool)
}


type defMsgServer struct {
	wgp   sync.WaitGroup
	stop  int32         		//停止标记
	cstop chan struct{}

	msgTyp    MsgType       		//消息类型
	netType   NetType
	handler   IMsgHandler  		    //消息处理器
	timeout   int                   //传输超时
	multiplex bool 		  		    //是否使用协程来处理MESSAGE
	writeChan int32                 //写通道消息缓存
}

func (r *defMsgServer) init(msgTyp MsgType,netType NetType, handler IMsgHandler) {
	if r.cstop !=nil{
		return
	}
	r.cstop = make(chan struct{})
	r.msgTyp = msgTyp
	r.netType = netType
	r.handler = handler
	r.writeChan = 100
	r.timeout = DefMsgQueTimeout
	r.wgp.Add(1)
}
//内部循环体使用,返回FALSE结束所有循环
func (r *defMsgServer) loop() bool {
	if r.stop == 0 && !loop(){
		r.Stop()
	}
	return r.stop == 0
}

//停止所有服务器
func (r *defMsgServer) Stop() {
	if !atomic.CompareAndSwapInt32(&r.stop, 0, 1) {
		return
	}
	close(r.cstop)
	r.wgp.Done()
	r.wgp.Wait()
}

func (r *defMsgServer) SetTimeout(t int) {
	if t >= 0 {
		r.timeout = t
	}
}
func (r *defMsgServer) GetTimeout() int {
	return r.timeout
}
func (r *defMsgServer) GetMsgType() MsgType {
	return r.msgTyp
}

func (r *defMsgServer) GetNetType() NetType {
	return r.netType
}

func (r *defMsgServer) GetHandler() IMsgHandler {
	return r.handler
}

func (r *defMsgServer) SetMultiplex(multiplex bool) {
	r.multiplex = multiplex
}

//IMsgQue 相当于socket.io中的socket,客户端连接信息
//各种服务器(TCP,UDP,WS)也使用该接口
type IMsgQue interface {
	Id() uint32

	LocalAddr() string
	RemoteAddr() string
	SetRealRemoteAddr(addr string)

	Stop()
	Available() bool
	IsProxy() bool

	Send(m *Message) (re bool)

	SetUser(user interface{})
	GetUser() interface{}
}



type defMsgQue struct {
	id uint32 //唯一标示
	stop    int32         //停止标记
	cwrite  chan *Message //写入通道
	lastTick int64        //最后有效行为时间戳

	handler     IMsgHandler
	msgServer   IMsgServer

	user           interface{} //玩家登陆后信息
	init      bool
	available bool
	multiplex bool 				//是否使用协程来处理MESSAGE

	realRemoteAddr string      //当使用代理是，需要特殊设置客户端真实IP
}


func (r *defMsgQue) Id() uint32 {
	return r.id
}

func (r *defMsgQue) SetUser(user interface{}) {
	r.user = user
}


func (r *defMsgQue) GetUser() interface{} {
	return r.user
}



//为FOR循环提供判断程序是不是还在执行，
func (r *defMsgQue) loop() bool {
	if r.stop == 0 && !r.msgServer.loop(){
		r.Stop()
	}
	return r.stop == 0
}


func (r *defMsgQue) isTimeout(tick *time.Timer) bool {
	timeout := r.msgServer.GetTimeout()
	timesTamp := time.Now().Unix()
	left := int(timesTamp - r.lastTick)
	if left < timeout || timeout == 0 {
		if timeout == 0 {
			tick.Reset(time.Second * time.Duration(DefMsgQueTimeout))
		} else {
			tick.Reset(time.Second * time.Duration(timeout-left))
		}
		return false
	}
	Logger.Debug("msgque close because timeout id:%v wait:%v timeout:%v", r.id, left, timeout)
	return true
}



func (r *defMsgQue) IsProxy() bool {
	return r.realRemoteAddr != ""
}

func (r *defMsgQue) SetRealRemoteAddr(addr string) {
	r.realRemoteAddr = addr
}



func (r *defMsgQue) Send(m *Message) (re bool) {
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
		r.Stop() //通道已满，直接关闭
	}

	return true
}

//
func (r *defMsgQue) Stop() {
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
func (r *defMsgQue) processMsg(msgque IMsgQue, msg *Message) bool {
	if r.multiplex {
		go r.processMsgTrue(msgque, msg)
	} else {
		return r.processMsgTrue(msgque, msg)
	}
	return true
}
func (r *defMsgQue) processMsgTrue(msgque IMsgQue, msg *Message) bool {
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
