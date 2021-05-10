package cosnet

import (
	"cosgo/logger"
	"net"
	"time"
)

type TcpServer struct {
	*NetServer
	listener net.Listener //监听
}

type TcpClient struct {
	TcpServer
}

func (s *TcpClient) Start() error {
	return nil
}

func NewTcpServer(address string, handler Handler) *TcpServer {
	return &TcpServer{
		NetServer: NewNetServer(address, handler, MsgTypeMsg, NetTypeTcp),
	}
}

func NewTcpClient(handler Handler) *TcpClient {
	s := &TcpClient{
		TcpServer: *NewTcpServer("", handler),
	}
	return s
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

func (s *TcpServer) Dial(address string) (sock Socket) {
	c, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		logger.Debug("connect to addr:%s failed err:%v", address, err)
	} else {
		sock = s.socket(c)
	}
	return
}

func (s *TcpServer) listen() {
	defer s.listener.Close()
	for !s.Done() {
		c, err := s.listener.Accept()
		if err != nil {
			logger.Error("tcp server accept failed:%v", err)
			break
		} else {
			s.GO(func() {
				s.socket(c)
			})
		}
	}
}

func (s *TcpServer) socket(conn net.Conn) Socket {
	sock := &TcpSocket{
		conn:      conn,
		NetSocket: NewSocket(s.Ctx(), s.handler),
	}
	if s.handler.Emit(HandlerEventTypeConnect, sock) {
		s.GO(sock.readMsg)
		s.GO(sock.writeMsg)
		logger.Debug("new socket Id:%d from Addr:%s", sock.id, sock.RemoteAddr())
	} else if sock.Close() {
		sock.conn.Close()
		sock = nil
	}
	return sock
}
