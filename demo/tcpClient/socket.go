package main

import (
	"cosgo/cosnet"
	"cosgo/logger"
	"cosgo/utils"
	"time"
)

var msg *cosnet.Message
var sockets = cosnet.NewSockets(&cosnet.HandlerDefault{}, 1024)
var handler *TcpHandler

func init() {
	msg = &cosnet.Message{Head: &cosnet.Header{Index: 1}}
	handler = &TcpHandler{}

}

func main() {
	address := "0.0.0.0:3100"
	for i := 1; i <= 1000; i++ {
		cosnet.NewTcpClient(handler, address)
	}
	scc := cosnet.SCC()
	scc.CGO(startSocketHeartbeat)
	scc.Wait()
}

func startSocketHeartbeat(stop chan struct{}) {
	t := time.Second * 5
	ticker := time.NewTimer(t)
	defer ticker.Stop()
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			utils.Try(heartbeat)
			ticker.Reset(t)
		}
	}
}

func heartbeat() {
	if sockets.Size() == 0 {
		return
	}
	sockets.Broadcast(msg.NewMsg(123, []byte("321")), nil)
}

type TcpHandler struct {
	cosnet.HandlerDefault
}

func (this *TcpHandler) Message(sock cosnet.Socket, msg *cosnet.Message) bool {
	logger.Debug("OnMessage:%+v", msg)
	return true
}
