package cosnet

import (
	"errors"
	"fmt"
	"github.com/hwcer/cosgo/logger"
	"io"
	"net"
	"net/url"
	"time"
)

type tcpSocket struct {
	conn   net.Conn
	agents *Agents
}

func (this *tcpSocket) Read(head []byte) (Message, error) {
	var (
		err error
		msg Message
	)
	_, err = io.ReadFull(this.conn, head)
	if err != nil {
		return nil, err
	}
	//logger.Debug("READ HEAD:%v", head)
	msg, err = this.agents.Handler.Parse(head)
	if err != nil {
		//logger.Debug("READ ERR:%v", err)
		return nil, err
	}
	if msg.Size() > 0 {
		data := make([]byte, msg.Size())
		_, err = io.ReadFull(this.conn, data)
		if err != nil {
			return nil, err
		}
		msg.Reset(data)
	}
	return msg, nil
}

func (this *tcpSocket) Write(msg Message) (err error) {
	var data []byte
	data, err = msg.Bytes()
	if err != nil {
		return err
	}
	var n int
	writeCount := 0
	for writeCount < len(data) {
		n, err = this.conn.Write(data[writeCount:])
		if err != nil {
			return err
		}
		writeCount += n
	}
	return nil
}

func (this *tcpSocket) Close() error {
	if this.conn != nil {
		return this.conn.Close()
	}
	return nil
}

func (this *tcpSocket) LocalAddr() net.Addr {
	if this.conn != nil {
		return this.conn.LocalAddr()
	}
	return nil
}

func (this *tcpSocket) RemoteAddr() net.Addr {
	if this.conn != nil {
		return this.conn.RemoteAddr()
	}
	return nil
}

type tcpServer struct {
	agents   *Agents
	address  *url.URL
	listener net.Listener
}

func NewTcpServer(agent *Agents, address *url.URL) (*tcpServer, error) {
	server := &tcpServer{agents: agent, address: address}
	if listener, err := net.Listen(NetworkTcp.String(), address.Host); err != nil {
		return nil, err
	} else {
		server.listener = listener
	}
	return server, nil
}

//Start listener.Accept
func (this *tcpServer) Start() {
	for !this.agents.scc.Stopped() {
		c, err := this.listener.Accept()
		if err != nil {
			logger.Debug("listener.Accept close")
			return
		}
		io := &tcpSocket{conn: c, agents: this.agents}
		this.agents.New(io, NetworkTcpServer)
	}
}

func (this *tcpServer) Close() error {
	return this.listener.Close()
}

//create client create
func NewTcpConnect(agents *Agents, address string) (*Socket, error) {
	conn, err := tryTcpConnect(address)
	if err != nil {
		return nil, err
	}
	io := &tcpSocket{conn: conn, agents: agents}
	return agents.New(io, NetworkTcpClient)
}

func tryTcpConnect(address string) (net.Conn, error) {
	for try := uint16(0); try <= Options.ClientReconnectMax; try++ {
		conn, err := net.DialTimeout("tcp", address, time.Second)
		if err == nil {
			return conn, nil
		} else {
			fmt.Printf("%v create error:%v\n", try, err)
			time.Sleep(time.Duration(Options.ClientReconnectTime))
		}
	}
	return nil, errors.New("Failed to create to Server")
}
