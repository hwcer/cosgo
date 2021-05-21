package cosnet

import (
	"github.com/hwcer/cosgo/logger"
	"net"
	"time"
)

type TcpServer struct {
	*NetServer
	listener net.Listener //监听
}

func NewTcpServer(address string, handler Handler) *TcpServer {
	return &TcpServer{
		NetServer: NewNetServer(address, handler, MsgTypeMsg, NetTypeTcp),
	}
}

func (s *TcpServer) Close() error {
	if !s.SCC.Close() {
		return nil
	}
	s.listener.Close()
	return s.SCC.Wait(time.Second * 10)
}

func (s *TcpServer) Start() error {
	if err := s.NetServer.Start(); err != nil {
		return err
	}
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	s.listener = listener
	s.GO(s.listen)
	return nil
}

func (s *TcpServer) listen() {
	//defer s.SCC.Cancel()
	for !s.Stopped() {
		c, err := s.listener.Accept()
		if err != nil {
			//logger.Error("tcp server accept failed:%v", err)
			break
		} else {
			go s.socket(c)
		}
	}
}

func (s *TcpServer) socket(conn net.Conn) Socket {
	sock := &TcpSocket{
		conn:      conn,
		NetSocket: NewSocket(s),
	}
	s.Emit(EventsTypeConnect, sock)
	s.GO(sock.readMsg)
	s.GO(sock.writeMsg)
	logger.Debug("new socket Id:%d from Addr:%s", sock.id, sock.RemoteAddr())
	return sock
}
