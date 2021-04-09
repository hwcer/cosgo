package main

import (
	"cosgo/cosnet"
	"cosgo/utils"
	"time"
)

var msg *cosnet.Message
var sockets = cosnet.NewSockets(&cosnet.HandlerDefault{}, 1024)

func init() {
	msg = &cosnet.Message{Head: &cosnet.Header{Index: 1}}
}

func main() {
	address := "0.0.0.0:3100"
	for i := 1; i <= 1000; i++ {
		cosnet.NewTcpClient(sockets, address)
	}
	scc := sockets.SCC()
	scc.CGO(startSocketHeartbeat)
	sockets.Wait()
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
		sockets.Close()
		return
	}
	sockets.Broadcast(msg.NewMsg(123, []byte("321")))
}
