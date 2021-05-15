package cosnet

import (
	"cosgo/logger"
	"net"
	"time"
)

type TcpClient struct {
	*TcpServer
}

func (s *TcpClient) Start() error {
	if s.address != "" {
		s.Connect(s.address)
	}
	return nil
}

func (s *TcpClient) Connect(address string) (sock Socket) {
	c, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		logger.Debug("connect to addr:%s failed err:%v", address, err)
	} else {
		sock = s.socket(c)
	}
	return
}

func NewTcpClient(address string, handler Handler) *TcpClient {
	s := &TcpClient{
		TcpServer: NewTcpServer(address, handler),
	}
	return s
}
