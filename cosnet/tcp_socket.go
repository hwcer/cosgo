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
	var (
		err error
		msg *message.Message
	)
	head := make([]byte, message.HeadSize)
	for !s.Stopped() {
		_, err = io.ReadFull(s.conn, head)
		if err != nil {
			if _, ok := err.(*net.OpError); !ok && err != io.EOF {
				logger.Debug("socket recv head err:%v", err)
			}
			break
		}
		msg, err = message.NewMsg(head)
		if err != nil {
			logger.Debug("socket parse head err:%v", err)
			break
		}
		if attachSize := msg.Attach.Size(); attachSize > 0 {
			attachByte := make([]byte, attachSize)
			_, err = io.ReadFull(s.conn, attachByte)
			if err != nil {
				logger.Debug("socket recv attach err:%v", err)
				break
			}
			err = msg.Attach.Parse(attachByte)
			if err != nil {
				logger.Debug("socket parse attach err:%v", err)
				break
			}
		}
		if msg.Head.Size > 0 {
			msg.Data = make([]byte, msg.Head.Size)
			_, err = io.ReadFull(s.conn, msg.Data)
			if err != nil {
				logger.Debug("socket recv data err:%v", err)
				break
			}
		}
		s.processMsg(s, msg)
		msg = nil
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
			logger.Error("socket write error,Id:%v err:%v", s.Id(), err)
			return
		}
		writeCount += n
	}

	s.KeepAlive()
}
