package main

import (
	"context"
	"cosgo/cosnet"
	"cosgo/logger"
	"cosgo/utils"
	"time"
)

var msg *cosnet.Message
var client *cosnet.TcpClient

func init() {
	msg = &cosnet.Message{Head: &cosnet.Header{Index: 1}}
	client = cosnet.NewTcpClient(&TcpHandler{})
}

func main() {
	address := "0.0.0.0:3100"
	for i := 1; i <= 1000; i++ {
		client.Dial(address)
	}
	client.CGO(startSocketHeartbeat)
	client.Wait()
}

func startSocketHeartbeat(ctx context.Context) {
	t := time.Second * 5
	ticker := time.NewTimer(t)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			utils.Try(heartbeat)
			ticker.Reset(t)
		}
	}
}

func heartbeat() {
	sockets := client.Sockets()
	if sockets.Size() == 0 {
		return
	}
	sockets.Broadcast(msg.NewMsg(123, []byte("321")), nil)
}

type TcpHandler struct {
	cosnet.HandlerDefault
}

func (this *TcpHandler) Message(ctx context.Context, sock cosnet.Socket, msg *cosnet.Message) bool {
	logger.Debug("OnMessage:%+v", msg)
	return true
}
