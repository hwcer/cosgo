package main

import (
	"github.com/hwcer/cosgo/cosnet"
	"github.com/hwcer/cosgo/cosnet/message"
)

var client *cosnet.TcpClient

func init() {
	message.SetAttachField("test", 4)
	msg := message.New(123, 1, message.ContentTypeNumber)
	msg.Attach.Set("test", 8)
	client = cosnet.NewTcpClient("", &TcpHandler{})
	client.On(cosnet.EventsTypeConnect, func(socket cosnet.Socket) {
		socket.Write(msg)
	})
}

func main() {
	address := "127.0.0.1:3100"
	for i := 1; i <= 2000; i++ {
		client.Connect(address)
	}
	client.Wait(0)
}

type TcpHandler struct {
}

func (this *TcpHandler) Message(sock cosnet.Socket, msg *message.Message) {
	//logger.Debug("OnMessage:%+v", msg)
	//i := utils.Random.Roll(1, 5)
	//<-time.After(time.Second * time.Duration(i))
	var v int32
	msg.Attach.Get("test", &v)
	msg.Attach.Set("test", v+1)
	sock.Write(msg)
}
