package cosnet

import (
	"github.com/hwcer/cosgo/library/logger"
	"net"
)

type debugSocket struct {
	agents *Agents
	handle func() (Message, error)
}

func (this *debugSocket) Read(head []byte) (Message, error) {
	return this.handle()
}

func (this *debugSocket) Write(msg Message) (err error) {
	logger.Debug("Debug Socket Write code:%v,size:%v,data:%v", msg.Code(), msg.Size(), string(msg.Data()))
	return nil
}

func (this *debugSocket) Close() error {
	return nil
}

func (this *debugSocket) LocalAddr() net.Addr {
	return nil
}

func (this *debugSocket) RemoteAddr() net.Addr {
	return nil
}

//create client create
func NewDebugSocket(agents *Agents, handle func() (Message, error)) (*Socket, error) {
	io := &debugSocket{agents: agents, handle: handle}
	return agents.New(io, NetworkTcpClient)
}
