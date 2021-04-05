package cosnet

import (
	"cosgo/logger"
	"errors"
	"sync"
	"sync/atomic"
	"time"
)

//NewSocketMgr cap初始容器大小
func NewSockets(cap int) *Sockets {
	sockets := &Sockets{
		dirty:  make(dirty, cap),
		slices: make([]Socket, 0, cap),
	}
	sockets.startHeartbeat()
	return sockets
}

type dirty map[int]bool

func (d dirty) get() (id int) {
	if len(d) == 0 {
		return
	}
	for id, _ = range d {
		break
	}
	delete(d, id)
	return
}

func (d dirty) add(id int) {
	d[id] = true
}

//socket 管理器
type Sockets struct {
	mu        sync.Mutex
	stop      int32
	seed      uint32 //ID 生成种子
	dirty     dirty
	slices    []Socket
	handler   Handler //消息处理器
	timestamp int64   //时间
	Multiplex bool    //是否使用协程来处理MESSAGE
}

//idPack 使用index生成ID
func (s *Sockets) idPack(index int) uint64 {
	s.seed++
	return uint64(index)<<32 | uint64(s.seed)
}

//idParse 返回idPack中的index
func (s *Sockets) idParse(id uint64) int {
	return int(id >> 32)
}

//创建一个新Socket
func (s *Sockets) New() *NetSocket {
	return NewNetSocket(s)
}

func (s *Sockets) Add(sock Socket) uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	var index int
	if index = s.dirty.get(); index > 0 {
		s.slices[index] = sock
	} else {
		index = len(s.slices)
		s.slices = append(s.slices, sock)
	}
	return s.idPack(index)
}
func (s *Sockets) Get(id uint64) Socket {
	index := s.idParse(id)
	if index >= len(s.slices) {
		return nil
	}
	if sock := s.slices[index]; sock.Id() == id {
		return sock
	} else {
		return nil
	}
}
func (s *Sockets) Delete(id uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	index := s.idParse(id)

	if index >= len(s.slices) || s.slices[index] == nil || s.slices[index].Id() != id {
		return
	}
	s.slices[index] = nil
	s.dirty.add(index)
}

//遍历
func (s *Sockets) Range(f func(Socket)) {
	for _, sock := range s.slices {
		if sock != nil {
			f(sock)
		}
	}
}

func (s *Sockets) Close() error {
	if !atomic.CompareAndSwapInt32(&s.stop, 0, 1) {
		return errors.New("server stoping")
	}
	return nil
}

func (s *Sockets) Stopped() bool {
	return s.stop == 1
}

//heartbeat 用来定时清理无效用户
func (s *Sockets) heartbeat() {
	for _, sock := range s.slices {
		if sock != nil {
			sock.timeout()
		}
	}
}

func (s *Sockets) startHeartbeat() {
	go func() {
		t := time.Millisecond * time.Duration(Config.Heartbeat)
		ticker := time.NewTimer(t)
		defer ticker.Stop()
		for !s.Stopped() {
			select {
			case <-ticker.C:
				s.timestamp += Config.Heartbeat
				s.heartbeat()
				ticker.Reset(t)
			}
		}
	}()
}

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
	Activity()
	timeout()
}

func NewNetSocket(sockets *Sockets) *NetSocket {
	sock := &NetSocket{
		cwrite:  make(chan *Message, Config.WriteChanSize),
		sockets: sockets,
	}
	return sock
}

type NetSocket struct {
	id             uint64        //唯一标示
	user           interface{}   //玩家登陆后信息
	stop           int32         //停止标记
	cwrite         chan *Message //写入通道
	handler        Handler
	sockets        *Sockets
	timestamp      int64  //最后有效行为时间戳
	realRemoteAddr string //当使用代理是，需要特殊设置客户端真实IP
}

func (s *NetSocket) timeout() {
	if s.sockets.timestamp-s.timestamp >= Config.ConnectTimeout {
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
		close(s.cwrite)
	}
	s.sockets.Delete(s.id)
	logger.Debug("socket Close Id:%d", s.id)
	return true
}

//判断连接是否关闭
func (s *NetSocket) Stopped() bool {
	if s.sockets.Stopped() {
		s.Close()
	}
	return s.stop > 0
}

func (s *NetSocket) IsProxy() bool {
	return s.realRemoteAddr != ""
}

func (s *NetSocket) Activity() {
	s.timestamp = s.sockets.timestamp
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

func (s *NetSocket) processMsg(msgque Socket, msg *Message) bool {
	s.Activity()
	if s.sockets.Multiplex {
		go s.processMsgTrue(msgque, msg)
	} else {
		return s.processMsgTrue(msgque, msg)
	}
	return true
}
func (s *NetSocket) processMsgTrue(sock Socket, msg *Message) bool {
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
