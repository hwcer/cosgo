package network

import (
	"io"
	"net"
	"sync/atomic"
	"time"
)
type tcpMsgServer struct {
	defMsgServer
	listener net.Listener //监听
}


func (r *tcpMsgServer) listen() {
	r.wgp.Add(1)
	defer r.wgp.Done()
	defer r.listener.Close()
	for r.loop() {
		c, err := r.listener.Accept()
		if err != nil {
			if stop == 0 && r.stop == 0 {
				Logger.Error("tcp server accept failed:%v",err)
			}
			break
		} else {
			Go(func() {
				msgque := newTcpMsgQue(r,c)
				if r.handler.OnNewMsgQue(msgque) {
					msgque.init = true
					msgque.available = true
					Go(func() {
						Logger.Debug("process read for msgque:%d", msgque.id)
						msgque.read()
						Logger.Debug("process read end for msgque:%d", msgque.id)
					})
					Go(func() {
						Logger.Debug("process write for msgque:%d", msgque.id)
						msgque.write()
						Logger.Debug("process write end for msgque:%d", msgque.id)
					})
				} else {
					msgque.Stop()
				}
			})
		}
	}
	r.Stop()
}


func NewTcpServer(addr string, msgTyp MsgType, handler IMsgHandler) (*tcpMsgServer,error) {
	listener, err := net.Listen("tcp", addr)
	if err !=nil{
		return nil,err
	}
	srv := &tcpMsgServer{
		listener: listener,
	}
	srv.defMsgServer.init(msgTyp,NetTypeTcp,handler)

	Go(func() {
		srv.listen()
	})
	return srv,nil
}




type tcpMsgQue struct {
	defMsgQue
	conn     net.Conn     //连接
	msgSrv *tcpMsgServer
	network  string
	address  string
	listener net.Listener //监听

}

func newTcpMsgQue(msgServer *tcpMsgServer,  conn net.Conn) *tcpMsgQue {
	msgque := tcpMsgQue{
		defMsgQue: defMsgQue{
			id:        atomic.AddUint32(&msgqueId, 1),
			cwrite:    make(chan *Message, 500),
			lastTick:  Timestamp,
			msgServer: msgServer,
		},
		conn: conn,
	}
	msgqueMapSync.Lock()
	msgqueMap[msgque.id] = &msgque
	msgqueMapSync.Unlock()
	Logger.Debug("new msgque id:%d from addr:%s", msgque.id, conn.RemoteAddr().String())
	return &msgque
}





func (r *tcpMsgQue) LocalAddr() string {
	if r.conn != nil {
		return r.conn.LocalAddr().String()
	} else if r.listener != nil {
		return r.listener.Addr().String()
	}
	return ""
}

func (r *tcpMsgQue) RemoteAddr() string {
	if r.realRemoteAddr != "" {
		return r.realRemoteAddr
	}
	if r.conn != nil {
		return r.conn.RemoteAddr().String()
	}
	return r.address
}

func (r *tcpMsgQue) readMsg() {
	headData := make([]byte, MsgHeadSize)
	var data []byte
	var head *MsgHead

	for r.loop() {
		if head == nil {
			_, err := io.ReadFull(r.conn, headData)
			if err != nil {
				if err != io.EOF {
					Logger.Debug("msgque:%v recv data err:%v", r.id, err)
				}
				break
			}
			if head = NewMsgHead(headData); head == nil {
				Logger.Debug("msgque:%v read msg head failed", r.id)
				break
			}
			if head.Len == 0 {
				if !r.processMsg(r, &Message{Head: head}) {
					Logger.Debug("msgque:%v process msg act:%v", r.id, head.Act)
					break
				}
				head = nil
			} else {
				data = make([]byte, head.Len)
			}
		} else {
			_, err := io.ReadFull(r.conn, data)
			if err != nil {
				Logger.Debug("msgque:%v recv data err:%v", r.id, err)
				break
			}
			if !r.processMsg(r, &Message{Head: head, Data: data}) {
				Logger.Debug("msgque:%v process msg act:%v ", r.id, head.Act)
				break
			}

			head = nil
			data = nil
		}
		r.lastTick = Timestamp
	}
}

func (r *tcpMsgQue) writeMsg() {
	var m *Message
	var data []byte
	writeCount := 0
	tick := time.NewTimer(time.Second * time.Duration(r.msgServer.GetTimeout()))
	for r.loop() || m != nil {
		if m == nil {
			select {
			case <-stopChanForGo:
			case m = <-r.cwrite:
				if m != nil {
					data = m.Bytes()
				}
			case <-tick.C:
				r.isTimeout(tick)
			}
		}

		if m == nil {
			continue
		}

		if writeCount < len(data) {
			n, err := r.conn.Write(data[writeCount:])
			if err != nil {
				Logger.Error("msgque write id:%v err:%v", r.id, err)
				break
			}
			writeCount += n
		}

		if writeCount == len(data) {
			writeCount = 0
			m = nil
		}
		r.lastTick = Timestamp
	}
	tick.Stop()
}

func (r *tcpMsgQue) read() {
	defer func() {
		if err := recover(); err != nil {
			Logger.Error("msgque read panic id:%v err:%v", r.id, err)
		}
		r.Stop()
	}()
	r.readMsg()
}

func (r *tcpMsgQue) write() {
	defer func() {
		if err := recover(); err != nil {
			Logger.Error("msgque write panic id:%v err:%v", r.id, err)
		}
		if r.conn != nil {
			r.conn.Close()
		}
		r.Stop()
	}()
	r.writeMsg()
}

