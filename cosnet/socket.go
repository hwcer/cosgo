package cosnet

import (
	"cosgo/logger"
	"cosgo/utils"
	"errors"
	"sync"
	"sync/atomic"
	"time"
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

func NewSocket(sockets *Sockets) (sock *NetSocket) {
	sock = &NetSocket{
		cwrite:    make(chan *Message, Config.WriteChanSize),
		sockets:   sockets,
		timestamp: sockets.timestamp,
	}
	sockets.Add(sock)
	return
}

//NewSocketMgr cap初始容器大小
func NewSockets(handler Handler, cap int) *Sockets {
	sockets := &Sockets{
		scc:     utils.NewSCC(),
		seed:    1,
		dirty:   newArrayMapDelIndex(cap),
		slices:  make([]Socket, cap, cap),
		handler: handler,
	}
	for i := 0; i < cap; i++ {
		sockets.dirty.Add(i)
	}
	sockets.startHeartbeat()
	return sockets
}

func newArrayMapDelIndex(cap int) *arrayMapDelIndex {
	return &arrayMapDelIndex{
		list:  make([]int, 0, cap),
		index: -1,
	}
}

//已经被删除的index
type arrayMapDelIndex struct {
	list  []int
	index int
}

func (this *arrayMapDelIndex) Add(val int) {
	this.index += 1
	if this.index < len(this.list) {
		this.list[this.index] = val
	} else {
		this.list = append(this.list, val)
	}
}

func (this *arrayMapDelIndex) Get() int {
	if this.index < 0 {
		return -1
	}
	val := this.list[this.index]
	this.list[this.index] = -1
	this.index -= 1
	return val
}

func (this *arrayMapDelIndex) Size() int {
	return this.index + 1
}

//socket 管理器
type Sockets struct {
	scc       *utils.SCC
	seed      uint32 //ID 生成种子
	mutex     sync.Mutex
	dirty     *arrayMapDelIndex
	slices    []Socket
	handler   Handler //消息处理器
	timestamp int64   //时间
	Multiplex bool    //是否使用协程来处理MESSAGE
}

//createSocketId 使用index生成ID
func (s *Sockets) createSocketId(index int) uint64 {
	s.seed++
	return uint64(index)<<32 | uint64(s.seed)
}

//parseSocketId 返回idPack中的index
func (s *Sockets) parseSocketId(id uint64) int {
	return int(id >> 32)
}

func (s *Sockets) SCC() *utils.SCC {
	return s.scc
}

//创建一个新NetSocket
func (s *Sockets) New() *NetSocket {
	return NewSocket(s)
}

func (s *Sockets) Add(sock Socket) error {
	if sock.Id() > 0 {
		return s.Reset(sock)
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var index = -1
	if index = s.dirty.Get(); index >= 0 {
		s.slices[index] = sock
	} else {
		index = len(s.slices)
		s.slices = append(s.slices, sock)
	}
	sock.init(s.createSocketId(index))
	return nil
}

//Del 删除
func (s *Sockets) Del(id uint64) {
	index := s.parseSocketId(id)
	if index >= len(s.slices) || s.slices[index] == nil || s.slices[index].Id() != id {
		return
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.slices[index] = nil
	s.dirty.Add(index)
}

//Get 获取
func (s *Sockets) Get(id uint64) Socket {
	index := s.parseSocketId(id)
	if index >= len(s.slices) {
		return nil
	}
	if sock := s.slices[index]; sock.Id() == id {
		return sock
	} else {
		return nil
	}
}

//Reset重设Socket 一般原始NetSocket被继承后，新SOCKET要使用Reset来保存新的Socket
func (s *Sockets) Reset(sock Socket) error {
	id := sock.Id()
	if id <= 0 {
		return errors.New("sockets reset id empty")
	}
	index := s.parseSocketId(id)
	if index >= len(s.slices) || s.slices[index] == nil || s.slices[index].Id() != id {
		return errors.New("socket reset but unequal")
	}
	s.slices[index] = sock
	return nil
}

//Size 当前socket数量
func (s *Sockets) Size() int {
	return len(s.slices) - s.dirty.Size()
}

//遍历
func (s *Sockets) Range(f func(Socket)) {
	for _, sock := range s.slices {
		if sock != nil {
			f(sock)
		}
	}
}

//Broadcast 广播,filter 过滤函数，如果不为nil且返回false则不对当期socket进行发送消息
func (s *Sockets) Broadcast(msg *Message, filter func(Socket) bool) {
	for _, sock := range s.slices {
		if sock == nil || (filter != nil && !filter(sock)) {
			continue
		}
		sock.Write(msg)
	}
}
func (s *Sockets) Wait() {
	s.scc.Wait()
}
func (s *Sockets) Close() error {
	return s.scc.Close()
}

func (s *Sockets) Stopped() bool {
	return s.scc.Stopped()
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
	s.scc.CGO(func(stop chan struct{}) {
		t := time.Millisecond * time.Duration(Config.Heartbeat)
		ticker := time.NewTimer(t)
		defer ticker.Stop()
		for !s.Stopped() {
			select {
			case <-stop:
				return
			case <-ticker.C:
				s.timestamp += Config.Heartbeat
				utils.Try(s.heartbeat, func(err interface{}) {
					logger.Error("startHeartbeat:%v", err)
				})
				ticker.Reset(t)
			}
		}
	})
}

type NetSocket struct {
	id             uint64        //唯一标示
	user           interface{}   //玩家登陆后信息
	stop           int32         //停止标记
	cwrite         chan *Message //写入通道
	sockets        *Sockets      //socket管理器
	indexes        int           //socket中的索引
	timestamp      int64         //最后有效行为时间戳
	realRemoteAddr string        //当使用代理是，需要特殊设置客户端真实IP
}

func (s *NetSocket) timeout() {
	if s.sockets.timestamp-s.timestamp >= Config.ConnectTimeout {
		s.Close()
	}
}
func (s *NetSocket) init(id uint64) {
	s.id = id
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
	s.sockets.Del(s.id)
	logger.Debug("Socket Close Id:%d", s.id)
	return true
}

//判断连接是否关闭
func (s *NetSocket) Stopped() bool {
	if s.stop == 0 && s.sockets.Stopped() {
		s.Close()
	}
	return s.stop > 0
}

func (s *NetSocket) IsProxy() bool {
	return s.realRemoteAddr != ""
}

func (s *NetSocket) KeepAlive() {
	s.timestamp = s.sockets.timestamp
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

func (s *NetSocket) processMsg(msgque Socket, msg *Message) bool {
	s.KeepAlive()
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
	return s.sockets.handler.OnMessage(sock, msg)
}
