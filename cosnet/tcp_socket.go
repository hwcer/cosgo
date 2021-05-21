package cosnet

import (
	"github.com/hwcer/cosgo/cosnet/message"
	"github.com/hwcer/cosgo/logger"
	"io"
	"net"
)

type TcpSocket struct {
	*NetSocket
	conn     net.Conn     //连接
	listener net.Listener //监听
}

func (s *TcpSocket) Close() bool {
	if !s.NetSocket.Close() {
		return false
	}
	s.conn.Close()
	s.server.Emit(EventsTypeDisconnect, s)
	return true
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
	defer s.Close()
	head := make([]byte, message.MsgHeadSize)
	for !s.Stopped() {
		_, err := io.ReadFull(s.conn, head)
		if err != nil {
			if _, ok := err.(*net.OpError); !ok && err != io.EOF {
				logger.Debug("socket:%v recv data err:%v", s.id, err)
			}
			break
		}
		msg, err := message.NewMsgHead(head)
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
		s.processMsg(s, msg)
	}
}

func (s *TcpSocket) writeMsg() {
	defer s.Close()
	for !s.Stopped() {
		select {
		case <-s.ctx.Done():
			return
		case m := <-s.cwrite:
			s.writeMsgTrue(m)
		}
	}
}

func (s *TcpSocket) writeMsgTrue(m *message.Message) {
	if m == nil {
		return
	}
	data := m.Bytes()
	writeCount := 0
	for !s.Stopped() && writeCount < len(data) {
		n, err := s.conn.Write(data[writeCount:])
		if err != nil {
			s.Close()
			logger.Error("socket write error,Id:%v err:%v", s.id, err)
			return
		}
		writeCount += n
	}

	s.KeepAlive()
}
