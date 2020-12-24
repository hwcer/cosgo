package cosnet

import (
	"cosgo/app"
	"cosgo/logger"
	"io"
	"net"
	"sync/atomic"
	"time"
)

type tcpServer struct {
	defServer
	listener net.Listener //监听
}

func (r *tcpServer) Start() (err error) {
	app.Go(func() {
		r.listen()
	})
	return
}

func (r *tcpServer) listen() {
	defer r.listener.Close()
	for !r.Done() {
		c, err := r.listener.Accept()
		if err != nil {
			logger.Error("tcp server accept failed:%v", err)
			break
		}
		app.Go(func() {
			sock := newTcpSocket(r, c)
			if r.handler.OnNewMsgQue(sock) {
				app.Go(func() {
					logger.Debug("process read for msgque:%d", sock.id)
					sock.read()
					logger.Debug("process read end for msgque:%d", sock.id)
				})
				app.Go(func() {
					logger.Debug("process write for msgque:%d", sock.id)
					sock.write()
					logger.Debug("process write end for msgque:%d", sock.id)
				})
			} else {
				sock.Close()
			}
		})
	}
}

func NewTcpServer(addr string, msgTyp MsgType, handler IMsgHandler) (*tcpServer, error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return nil, err
	}
	srv := &tcpServer{
		defServer: defServer{msgTyp: msgTyp, netType: NetTypeTcp, handler: handler},
		listener:  listener,
	}
	return srv, nil
}

type tcpSocket struct {
	defSocket
	conn     net.Conn //连接
	address  string
	listener net.Listener //监听
}

func newTcpSocket(srv *tcpServer, conn net.Conn) *tcpSocket {
	sock := &tcpSocket{
		defSocket: defSocket{
			id:        atomic.AddUint32(&msgqueId, 1),
			cwrite:    make(chan *Message, Config.WriteChanSize),
			heartbeat: timestamp,
			server:    srv,
		},
		conn: conn,
	}
	msgqueMapSync.Lock()
	msgqueMap[sock.id] = sock
	msgqueMapSync.Unlock()
	logger.Debug("new socket Id:%d from Addr:%s", sock.id, conn.RemoteAddr().String())
	return sock
}

func (r *tcpSocket) LocalAddr() string {
	if r.conn != nil {
		return r.conn.LocalAddr().String()
	} else if r.listener != nil {
		return r.listener.Addr().String()
	}
	return ""
}

func (r *tcpSocket) RemoteAddr() string {
	if r.realRemoteAddr != "" {
		return r.realRemoteAddr
	}
	if r.conn != nil {
		return r.conn.RemoteAddr().String()
	}
	return r.address
}

func (r *tcpSocket) readMsg() {
	defer r.conn.Close()
	headData := make([]byte, MsgHeadSize)
	var data []byte
	var head *MsgHead

	for !r.Done() {
		if head == nil {
			_, err := io.ReadFull(r.conn, headData)
			if err != nil {
				if err != io.EOF {
					logger.Debug("msgque:%v recv data err:%v", r.id, err)
				}
				break
			}
			if head = NewMsgHead(headData); head == nil {
				logger.Debug("msgque:%v read msg head failed", r.id)
				break
			}
			if head.Len == 0 {
				if !r.processMsg(r, &Message{Head: head}) {
					logger.Debug("msgque:%v process msg act:%v", r.id, head.Proto)
					break
				}
				head = nil
			} else {
				data = make([]byte, head.Len)
			}
		} else {
			_, err := io.ReadFull(r.conn, data)
			if err != nil {
				logger.Debug("msgque:%v recv data err:%v", r.id, err)
				break
			}
			if !r.processMsg(r, &Message{Head: head, Data: data}) {
				logger.Debug("msgque:%v process msg act:%v ", r.id, head.Proto)
				break
			}

			head = nil
			data = nil
		}
		r.heartbeat = timestamp
	}
}

func (r *tcpSocket) writeMsg() {
	var m *Message
	var data []byte
	writeCount := 0
	tick := time.NewTimer(time.Millisecond * time.Duration(Config.ConnectHeartbeat))
	for !r.Done() || m != nil {
		if m == nil {
			select {
			case m = <-r.cwrite:
				if m != nil {
					data = m.Bytes()
				}
			case <-tick.C:
				if r.isTimeout(tick) {
					r.Close()
				}
			}
		}

		if m == nil {
			continue
		}

		if writeCount < len(data) {
			n, err := r.conn.Write(data[writeCount:])
			if err != nil {
				logger.Error("msgque write Id:%v err:%v", r.id, err)
				break
			}
			writeCount += n
		}

		if writeCount == len(data) {
			writeCount = 0
			m = nil
		}
		r.heartbeat = timestamp
	}
	tick.Stop()
}

func (r *tcpSocket) read() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("msgque read panic Id:%v err:%v", r.id, err)
		}
		r.Close()
	}()
	r.readMsg()
}

func (r *tcpSocket) write() {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("msgque write panic Id:%v err:%v", r.id, err)
		}
		if r.conn != nil {
			r.conn.Close()
		}
		r.Close()
	}()
	r.writeMsg()
}
