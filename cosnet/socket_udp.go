package cosnet

import (
	"context"
	"errors"
	"github.com/hwcer/cosgo/utils"
	"net"
	"net/url"
	"sync"
)

type udpSocket struct {
	addr   *net.UDPAddr
	conn   *net.UDPConn //连接
	cread  chan []byte  //写入通道
	agents *Agents
}

func (this *udpSocket) Read(head []byte) (Message, error) {
	var data []byte
	select {
	case data = <-this.cread:
	}
	if data == nil {
		return nil, errors.New("udp conn closed")
	}
	msg := this.agents.Handler.New()
	msg.Reset(data)
	return msg, nil
}

func (this *udpSocket) Write(msg Message) (err error) {
	var data []byte
	data, err = msg.Bytes()
	if err != nil {
		return err
	}
	_, err = this.conn.WriteToUDP(data, this.addr)
	return err
}

func (this *udpSocket) Close() error {
	utils.Try(func() {
		if this.cread != nil {
			close(this.cread)
		}
	})
	//UDP conn 所有连接共享，一个客户端断开，不能直接关闭conn
	//return this.conn.Close()
	return nil
}

func (this *udpSocket) LocalAddr() net.Addr {
	if this.conn != nil {
		return this.conn.LocalAddr()
	}
	return nil
}

func (this *udpSocket) RemoteAddr() net.Addr {
	return this.addr
}

type udpMsg struct {
	addr *net.UDPAddr
	data []byte
}

type udpServer struct {
	conn       *net.UDPConn
	address    *url.URL
	udpMsgChan chan *udpMsg
	mu         sync.Mutex
	dict       map[string]*udpSocket
	agents     *Agents
}

func (this *udpServer) Close() error {
	utils.Try(func() {
		this.conn.Close()
	})
	return nil
}

func (this *udpServer) Start() {
	data := make([]byte, 1<<16)
	for !this.agents.scc.Stopped() {
		n, addr, err := this.conn.ReadFromUDP(data)
		if err != nil {
			if err.(net.Error).Timeout() {
				continue
			} else {
				break
			}
		}

		if n <= 0 {
			continue
		}
		this.udpMsgChan <- &udpMsg{
			addr: addr,
			data: data,
		}
	}
}

//消息处理器
func (this *udpServer) handler(ctx context.Context) {
	for !this.agents.scc.Stopped() {
		select {
		case <-ctx.Done():
			return
		case m := <-this.udpMsgChan:
			this.accept(m)
		}
	}
}

func (this *udpServer) accept(m *udpMsg) {
	socket, err := this.Socket(m.addr)
	if err != nil {
		return //TODO ERROR
	}
	select {
	case socket.cread <- m.data:
	default:
		//TODO
	}
}
func (this *udpServer) remove(socket *Socket) {
	if !socket.HasType(NetworkUdpServer) {
		return
	}
	if s, ok := socket.io.(*udpSocket); ok {
		this.mu.Lock()
		delete(this.dict, s.addr.String())
		this.mu.Unlock()
	}
}

//Socket 获取并自动创建udpSocket
func (this *udpServer) Socket(addr *net.UDPAddr) (*udpSocket, error) {
	this.mu.Lock()
	defer this.mu.Unlock()
	k := addr.String()
	if socket := this.dict[k]; socket == nil {
		socket = &udpSocket{conn: this.conn, addr: addr, cread: make(chan []byte, 1024)}
		_, err := this.agents.New(socket, NetworkUdpServer)
		if err != nil {
			return nil, err
		}
		this.dict[k] = socket
		return socket, nil
	} else {
		return socket, nil
	}
}

func NewUdpServer(agent *Agents, address *url.URL) (Server, error) {
	naddr, err := net.ResolveUDPAddr(address.Scheme, address.Host)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP(address.Scheme, naddr)
	if err != nil {
		return nil, err
	}
	if err = conn.SetReadBuffer(1 << 24); err != nil {
		return nil, err
	}
	if err = conn.SetWriteBuffer(1 << 24); err != nil {
		return nil, err
	}
	srv := &udpServer{conn: conn, agents: agent, address: address}
	srv.dict = make(map[string]*udpSocket)
	srv.udpMsgChan = make(chan *udpMsg, 64)
	for i := 0; i < 10; i++ {
		agent.scc.CGO(srv.handler)
	}

	agent.On(EventTypeDisconnect, srv.remove)
	return srv, nil
}
