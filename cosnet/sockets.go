package cosnet

import (
	"context"
	"cosgo/utils"
	"sync"
	"time"
)

//NewSockets socket管理器 cap初始容器大小
func NewSockets(handler Handler, cap int) *Sockets {
	sockets := &Sockets{
		seed:    1,
		dirty:   newArrayMapDelIndex(cap),
		slices:  make([]Socket, cap, cap),
		handler: handler,
	}
	for i := cap - 1; i >= 0; i-- {
		sockets.dirty.Add(i)
	}
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
	seed      uint32 //ID 生成种子
	mutex     sync.Mutex
	dirty     *arrayMapDelIndex
	slices    []Socket
	handler   Handler //消息处理器
	heartbeat int
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

func (s *Sockets) Add(sock Socket) uint64 {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var index = -1
	if index = s.dirty.Get(); index >= 0 {
		s.slices[index] = sock
	} else {
		index = len(s.slices)
		s.slices = append(s.slices, sock)
	}
	return s.createSocketId(index)
}

//Del 删除
func (s *Sockets) Del(id uint64) bool {
	index := s.parseSocketId(id)
	if index >= len(s.slices) || s.slices[index] == nil || s.slices[index].Id() != id {
		return true
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.slices[index] = nil
	s.dirty.Add(index)
	return true
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

func (s *Sockets) Start(ctx context.Context) {
	t := time.Millisecond * time.Duration(Config.SocketHeartbeat)
	ticker := time.NewTimer(t)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			utils.Try(s.clearSocket)
			ticker.Reset(t)
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

//heartbeat 用来定时清理无效用户
func (s *Sockets) clearSocket() {
	for _, sock := range s.slices {
		if sock != nil {
			sock.Heartbeat()
		}
	}
}
