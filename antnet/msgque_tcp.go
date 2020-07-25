package antnet

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

type tcpMsgQue struct {
	msgQue
	wait     sync.WaitGroup
	conn     net.Conn     //连接
	listener net.Listener //监听
	network  string
	address  string
}

func (r *tcpMsgQue) GetNetType() NetType {
	return NetTypeTcp
}
func (r *tcpMsgQue) Stop() {
	if atomic.CompareAndSwapInt32(&r.stop, 0, 1) {
		Go(func() {
			if r.init {
				r.handler.OnDelMsgQue(r)
			}
			r.available = false
			r.baseStop()
		})
	}
}

func (r *tcpMsgQue) IsStop() bool {
	if r.stop == 0 {
		if IsStop() {
			r.Stop()
		}
	}
	return r.stop == 1
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

	for !r.IsStop() {
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
	tick := time.NewTimer(time.Second * time.Duration(r.timeout))
	for !r.IsStop() || m != nil {
		if m == nil {
			select {
			case <-stopChanForGo:
			case m = <-r.cwrite:
				if m != nil {
					data = m.Bytes()
				}
			case <-tick.C:
				if r.isTimeout(tick) {
					r.Stop()
				}
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
		r.wait.Done()
		if err := recover(); err != nil {
			Logger.Error("msgque read panic id:%v err:%v", r.id, err.(error))
		}
		r.Stop()
	}()

	r.wait.Add(1)
	r.readMsg()
}

func (r *tcpMsgQue) write() {
	defer func() {
		r.wait.Done()
		if err := recover(); err != nil {
			Logger.Error("msgque write panic id:%v err:%v", r.id, err.(error))
		}
		if r.conn != nil {
			r.conn.Close()
		}
		r.Stop()
	}()
	r.wait.Add(1)
	r.writeMsg()
}

func (r *tcpMsgQue) listen() {
	defer r.listener.Close()
	for !r.IsStop() {
		c, err := r.listener.Accept()
		if err != nil {
			if stop == 0 && r.stop == 0 {
				Logger.Error("accept failed msgque:%v err:%v", r.id, err)
			}
			break
		} else {
			Go(func() {
				msgque := newTcpAccept(c, r.msgTyp, r.handler)
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

func newTcpAccept(conn net.Conn, msgtyp MsgType, handler IMsgHandler) *tcpMsgQue {
	msgque := tcpMsgQue{
		msgQue: msgQue{
			id:       atomic.AddUint32(&msgqueId, 1),
			cwrite:   make(chan *Message, 500),
			msgTyp:   msgtyp,
			handler:  handler,
			timeout:  DefMsgQueTimeout,
			connTyp:  ConnTypeAccept,
			lastTick: Timestamp,
		},
		conn: conn,
	}
	msgqueMapSync.Lock()
	msgqueMap[msgque.id] = &msgque
	msgqueMapSync.Unlock()
	Logger.Debug("new msgque id:%d from addr:%s", msgque.id, conn.RemoteAddr().String())
	return &msgque
}

func newTcpListen(listener net.Listener, msgtyp MsgType, handler IMsgHandler, addr string) *tcpMsgQue {
	msgque := tcpMsgQue{
		msgQue: msgQue{
			id:      atomic.AddUint32(&msgqueId, 1),
			msgTyp:  msgtyp,
			handler: handler,
			connTyp: ConnTypeListen,
		},
		listener: listener,
	}

	msgqueMapSync.Lock()
	msgqueMap[msgque.id] = &msgque
	msgqueMapSync.Unlock()
	Logger.Debug("new tcp listen id:%d addr:%s", msgque.id, addr)
	return &msgque
}
