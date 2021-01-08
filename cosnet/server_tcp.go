package cosnet

import (
	"cosgo/logger"
	"io"
	"net"
)

func NewTcpServer(addr string, msgTyp MsgType, handler MsgHandler) (*TcpServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := &TcpServer{
		NetServer: NewNetServer(msgTyp, handler, NetTypeTcp),
		listener:  listener,
	}
	return srv, nil
}

type TcpSocket struct {
	*NetSocket
	conn     net.Conn //连接
	address  string
	listener net.Listener //监听
}

type TcpServer struct {
	*NetServer
	listener net.Listener //监听
}

func (s *TcpServer) Start() (err error) {
	s.Go(s.listen)
	return
}

func (r *TcpServer) listen() {
	defer r.stop.Close()
	defer r.listener.Close()
	for !r.Stoped() {
		c, err := r.listener.Accept()
		if err != nil {
			logger.Error("tcp server accept failed:%v", err)
			break
		} else {
			r.Go(func() { r.socket(c) })
		}
	}
}

func (r *TcpServer) socket(conn net.Conn) {
	sock := &TcpSocket{
		conn:      conn,
		NetSocket: NewNetSocket(r),
	}
	if r.handler.OnConnect(sock) {
		r.Go(sock.read)
		r.Go(sock.write)
		r.sockets.Add(sock)
		logger.Debug("new socket Id:%d from Addr:%s", sock.id, conn.RemoteAddr().String())
	} else if sock.Close() {
		sock.conn.Close()
	}
}

func (r *TcpSocket) LocalAddr() string {
	if r.conn != nil {
		return r.conn.LocalAddr().String()
	} else if r.listener != nil {
		return r.listener.Addr().String()
	}
	return ""
}

func (r *TcpSocket) RemoteAddr() string {
	if r.realRemoteAddr != "" {
		return r.realRemoteAddr
	}
	if r.conn != nil {
		return r.conn.RemoteAddr().String()
	}
	return r.address
}

func (r *TcpSocket) readMsg() {
	headData := make([]byte, MsgHeadSize)
	var data []byte
	var head *MsgHead

	for !r.Stoped() {
		if head == nil {
			_, err := io.ReadFull(r.conn, headData)
			if err != nil {
				if err != io.EOF {
					logger.Debug("socket:%v recv data err:%v", r.id, err)
				}
				break
			}
			if head = NewMsgHead(headData); head == nil {
				logger.Debug("socket:%v read msg head failed", r.id)
				break
			}
			if head.Len == 0 {
				if !r.processMsg(r, &Message{Head: head}) {
					logger.Debug("socket:%v process msg act:%v", r.id, head.Proto)
					break
				}
				head = nil
			} else {
				data = make([]byte, head.Len)
			}
		} else {
			_, err := io.ReadFull(r.conn, data)
			if err != nil {
				logger.Debug("socket:%v recv data err:%v", r.id, err)
				break
			}
			if !r.processMsg(r, &Message{Head: head, Data: data}) {
				logger.Debug("socket:%v process msg act:%v ", r.id, head.Proto)
				break
			}

			head = nil
			data = nil
		}
	}
}

func (r *TcpSocket) writeMsg() {
	var m *Message
	var data []byte
	writeCount := 0
	for !r.Stoped() || m != nil {
		if m == nil {
			select {
			case m = <-r.cwrite:
				if m != nil {
					data = m.Bytes()
				}
			case <-r.ticker.C:
				if r.timeout() {
					return
				}
			}
		}

		if m == nil {
			continue
		}

		if writeCount < len(data) {
			n, err := r.conn.Write(data[writeCount:])
			if err != nil {
				logger.Error("socket write error,Id:%v err:%v", r.id, err)
				break
			}
			writeCount += n
		}

		if writeCount == len(data) {
			writeCount = 0
			m = nil
		}
		r.heartbeat = r.server.Timestamp()
	}
}

func (r *TcpSocket) read() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("socket read panic Id:%v err:%v", r.id, err)
		}
		r.Close()
	}()
	r.readMsg()
}

func (r *TcpSocket) write() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("socket write panic Id:%v err:%v", r.id, err)
		}
		if r.conn != nil {
			r.conn.Close()
		}
		r.Close()
	}()

	r.tickerStart()
	defer r.tickerStop()
	r.writeMsg()
}
