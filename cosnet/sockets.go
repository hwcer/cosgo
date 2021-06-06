package cosnet

import (
	"context"
	"github.com/hwcer/cosgo/cosmap"
	"github.com/hwcer/cosgo/cosnet/message"
	"github.com/hwcer/cosgo/utils"
	"time"
)

//NewSockets socket管理器 cap初始容器大小
func NewSockets(handler Handler, cap int) *Sockets {
	sockets := &Sockets{
		Array:   cosmap.NewArray(cap),
		handler: handler,
	}
	return sockets
}

//socket 管理器
type Sockets struct {
	*cosmap.Array
	handler   Handler //消息处理器
	heartbeat int
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
func (s *Sockets) Broadcast(msg *message.Message, filter func(Socket) bool) {
	s.Array.Range(func(i interface{}) {
		sock := i.(Socket)
		if filter(sock) {
			sock.Write(msg)
		}
	})
}

//heartbeat 用来定时清理无效用户
func (s *Sockets) clearSocket() {
	s.Array.Range(func(i interface{}) {
		i.(Socket).Heartbeat()
	})
}
