package cosnet

import (
	"cosgo/logger"
	"io"
	"net"
)

func NewTcpServer(addr string, handler Handler) (*TcpServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := &TcpServer{
		NetServer: NewNetServer(MsgTypeMsg, handler, NetTypeTcp),
		listener:  listener,
	}
	return srv, nil
}

type TcpSocket struct {
	*NetSocket
	Conn     net.Conn     //连接
	Listener net.Listener //监听
}

type TcpServer struct {
	*NetServer
	listener net.Listener //监听
}

func (s *TcpServer) Start() (err error) {
	Go(s.wgp, s.listen)
	return
}

func (s *TcpServer) listen() {
	defer s.listener.Close()
	for !s.Stoped() {
		c, err := s.listener.Accept()
		if err != nil {
			logger.Error("tcp server accept failed:%v", err)
			break
		} else {
			SafeGo(s.wgp, func() { s.socket(c) })
		}
	}
}

func (s *TcpServer) socket(conn net.Conn) {
	sock := &TcpSocket{
		Conn:      conn,
		NetSocket: NewNetSocket(s),
	}
	if s.handler.OnConnect(sock) {
		SafeGo(s.wgp, sock.readMsg)
		SafeGo(s.wgp, sock.writeMsg)
		s.sockets.Add(sock)
		logger.Debug("new socket Id:%d from Addr:%s", sock.id, conn.RemoteAddr().String())
	} else if sock.Close() {
		sock.Conn.Close()
	}
}

func (s *TcpSocket) LocalAddr() string {
	if s.Conn != nil {
		return s.Conn.LocalAddr().String()
	} else if s.Listener != nil {
		return s.Listener.Addr().String()
	}
	return ""
}

func (s *TcpSocket) RemoteAddr() string {
	if s.realRemoteAddr != "" {
		return s.realRemoteAddr
	}
	if s.Conn != nil {
		return s.Conn.RemoteAddr().String()
	}
	return ""
}

func (s *TcpSocket) readMsg() {
	defer s.Close()

	headData := make([]byte, MsgHeadSize)
	var data []byte

	for !s.Stoped() {
		_, err := io.ReadFull(s.Conn, headData)
		if err != nil {
			if err != io.EOF {
				logger.Debug("socket:%v recv data err:%v", s.id, err)
			}
			break
		}
		head := NewMsgHead(headData)
		if head == nil {
			logger.Debug("socket:%v read msg head failed", s.id)
			break
		}
		if head.Len > 0 {
			data = make([]byte, head.Len)
			_, err := io.ReadFull(s.Conn, data)
			if err != nil {
				logger.Debug("socket:%v recv data err:%v", s.id, err)
				break
			}
		}
		if !s.processMsg(s, &Message{Head: head, Data: data}) {
			logger.Debug("socket:%v process msg act:%v ", s.id, head.Proto)
			break
		}
		data = nil
	}
}

func (s *TcpSocket) writeMsg() {
	defer func() {
		if s.Conn != nil {
			s.Conn.Close()
		}
		s.Close()
	}()

	for !s.Stoped() {
		select {
		case m := <-s.cwrite:
			if !s.writeMsgTrue(m) {
				return
			}
		case <-s.ticker.C:
			if s.timeout() {
				return
			}
		}
	}
}

func (s *TcpSocket) writeMsgTrue(m *Message) bool {
	if m == nil {
		return false
	}
	data := m.Bytes()
	writeCount := 0
	for !s.Stoped() && writeCount < len(data) {
		n, err := s.Conn.Write(data[writeCount:])
		if err != nil {
			logger.Error("socket write error,Id:%v err:%v", s.id, err)
			return false
		}
		writeCount += n
	}
	s.heartbeat = s.server.Timestamp()
	return true
}
