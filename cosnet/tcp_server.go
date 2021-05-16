package cosnet

import (
	"github.com/hwcer/cosgo/logger"
	"net"
)

type TcpServer struct {
	*NetServer
	listener net.Listener //监听
}

func (s *TcpServer) Close() error {
	return s.SCC.Close(func() {
		if s.listener != nil {
			s.listener.Close()
		}
	})
}

//Shutdown
func (s *TcpServer) Shutdown() {
	go s.Close()
}

func NewTcpServer(address string, handler Handler) *TcpServer {
	return &TcpServer{
		NetServer: NewNetServer(address, handler, MsgTypeMsg, NetTypeTcp),
	}
}

func (s *TcpServer) Start() error {
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	s.listener = listener
	s.GO(s.listen)
	return nil
}

func (s *TcpServer) listen() {
	defer s.Shutdown()
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
		NetSocket: NewSocket(s.handler),
	}
	if s.handler.Emit(HandlerEventTypeConnect, sock) {
		s.CGO(sock.readMsg)
		s.CGO(sock.writeMsg)
		logger.Debug("new socket Id:%d from Addr:%s", sock.id, sock.RemoteAddr())
	} else if sock.Close() {
		sock.conn.Close()
		sock = nil
	}
	return sock
}
