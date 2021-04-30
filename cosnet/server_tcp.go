package cosnet

import (
	"cosgo/logger"
	"io"
	"net"
	"time"
)

func NewTcpServer(address string, handler Handler) *TcpServer {
	return &TcpServer{
		NetServer: NewNetServer(address, handler, MsgTypeMsg, NetTypeTcp),
	}
}

type TcpServer struct {
	*NetServer
	listener net.Listener //监听
}
type TcpSocket struct {
	*NetSocket
	conn     net.Conn     //连接
	listener net.Listener //监听
}

func (s *TcpServer) start() error {
	//if err := s.NetServer.start(); err != nil {
	//	return err
	//}
	listener, err := net.Listen("tcp", s.address)
	if err != nil {
		return err
	}
	s.handler.On(HandlerEventTypeConnect, s.sockets.Add)
	s.handler.On(HandlerEventTypeDisconnect, s.sockets.Del)
	s.listener = listener
	scc.GO(s.listen)
	return nil
}

func (s *TcpServer) listen() {
	defer s.listener.Close()
	for !scc.Stopped() {
		c, err := s.listener.Accept()
		if err != nil {
			logger.Error("tcp server accept failed:%v", err)
			break
		} else {
			go NewTcpSocket(s.handler, c)
		}
	}
}

func (s *TcpSocket) LocalAddr() string {
	if s.conn != nil {
		return s.conn.LocalAddr().String()
	} else if s.listener != nil {
		return s.listener.Addr().String()
	}
	return ""
}

func (s *TcpSocket) RemoteAddr() string {
	if s.realRemoteAddr != "" {
		return s.realRemoteAddr
	}
	if s.conn != nil {
		return s.conn.RemoteAddr().String()
	}
	return ""
}

func (s *TcpSocket) readMsg() {
	//defer s.Close()
	head := make([]byte, MsgHeadSize)
	for !s.Stopped() {
		_, err := io.ReadFull(s.conn, head)
		if err != nil {
			if _, ok := err.(*net.OpError); !ok && err != io.EOF {
				logger.Debug("socket:%v recv data err:%v", s.id, err)
			}
			break
		}
		msg, err := NewMsg(head)
		if err != nil {
			logger.Debug("socket:%v read msg msg failed:%v", err)
			break
		}
		if msg.Head.Size > 0 {
			msg.Data = make([]byte, msg.Head.Size)
			_, err := io.ReadFull(s.conn, msg.Data)
			if err != nil {
				logger.Debug("socket:%v recv data err:%v", s.id, err)
				break
			}
		}
		if !s.processMsg(s, msg) {
			logger.Debug("socket:%v process msg act:%v ", s.id, msg.Head.Proto)
			break
		}
	}
}

func (s *TcpSocket) writeMsg() {
	defer func() {
		if s.conn != nil {
			s.conn.Close()
		}
		s.Close()
		s.handler.Emit(HandlerEventTypeDisconnect, s)
	}()

	for !s.Stopped() {
		select {
		case m := <-s.cwrite:
			if m != nil && !s.writeMsgTrue(m) {
				return
			}
		}
	}
}

func (s *TcpSocket) writeMsgTrue(m *Message) bool {
	data := m.Bytes()
	writeCount := 0
	for !s.Stopped() && writeCount < len(data) {
		n, err := s.conn.Write(data[writeCount:])
		if err != nil {
			logger.Error("socket write error,Id:%v err:%v", s.id, err)
			return false
		}
		writeCount += n
	}
	s.KeepAlive()
	return true
}

//客户端
func NewTcpClient(handler Handler, address string) (sock Socket) {
	c, err := net.DialTimeout("tcp", address, time.Second)
	if err != nil {
		logger.Debug("connect to addr:%s failed err:%v", address, err)
	} else {
		sock = NewTcpSocket(handler, c)
	}
	return
}

func NewTcpSocket(handler Handler, conn net.Conn) Socket {
	sock := &TcpSocket{
		conn:      conn,
		NetSocket: NewSocket(handler),
	}
	if handler.Emit(HandlerEventTypeConnect, sock) {
		scc.GO(sock.readMsg)
		scc.GO(sock.writeMsg)
		logger.Debug("new socket Id:%d from Addr:%s", sock.id, sock.RemoteAddr())
	} else if sock.Close() {
		sock.conn.Close()
		sock = nil
	}
	return sock
}
