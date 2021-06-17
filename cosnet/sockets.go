package cosnet

import (
	"context"
	"github.com/hwcer/cosgo/cosnet/message"
	"github.com/hwcer/cosgo/storage"
	"github.com/hwcer/cosgo/utils"
	"time"
)

//NewSockets socket管理器 cap初始容器大小
func NewSockets(handler Handler, cap int) *Sockets {
	sockets := &Sockets{
		Array:   storage.NewArray(cap),
		handler: handler,
	}
	sockets.Array.Multiplex = false
	return sockets
}

//socket 管理器
type Sockets struct {
	*storage.Array
	handler   Handler //消息处理器
	heartbeat int
}

func (s *Sockets) Get(id uint64) (Socket, bool) {
	if v, ok := s.Array.Dataset(id); !ok {
		return nil, false
	} else if v2, ok2 := v.(Socket); ok2 {
		return v2, true
	} else {
		return nil, false
	}
}

//func (s *Sockets) Set(sock Socket) uint64 {
//	return s.Array.SetDataset(sock.(storage.ArrayDataset))
//}

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
func (s *Sockets) Broadcast(msg *message.Message, filter func(Socket) bool) {
	s.Array.Range(func(v storage.ArrayDataset) bool {
		sock := v.(Socket)
		if filter(sock) {
			sock.Write(msg)
		}
		return true
	})
}

//heartbeat 用来定时清理无效用户
func (s *Sockets) clearSocket() {
	s.Array.Range(func(v storage.ArrayDataset) bool {
		v.(Socket).Heartbeat()
		return true
	})
}
