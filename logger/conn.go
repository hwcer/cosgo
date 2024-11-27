package logger

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

func NewConn(network, address string) *Conn {
	return &Conn{Network: network, Address: address}
}

type Conn struct {
	sync.Mutex
	Network     string `json:"network"`
	Address     string `json:"address"`
	Reconnect   bool   `json:"reconnect"`
	Format      func(*Message) string
	innerWriter io.WriteCloser
	illNetFlag  bool //网络异常标记
}

func (c *Conn) Name() string {
	return c.Network + "://" + c.Address
}

func (c *Conn) Init() error {
	if c.innerWriter != nil {
		_ = c.innerWriter.Close()
		c.innerWriter = nil
	}
	return nil
}

func (c *Conn) Write(msg *Message) (err error) {
	if c.needToConnectOnMsg() {
		err = c.connect()
		if err != nil {
			return
		}
		c.illNetFlag = false
	}
	//网络异常时，消息发出
	if !c.illNetFlag {
		err = c.println(msg)
		//网络异常，通知处理网络的go程自动重连
		if err != nil {
			c.illNetFlag = true
		}
	}
	return
}

func (c *Conn) Close() {
	if c.innerWriter != nil {
		_ = c.innerWriter.Close()
	}
}

func (c *Conn) connect() error {
	if c.innerWriter != nil {
		_ = c.innerWriter.Close()
		c.innerWriter = nil
	}
	addrs := strings.Split(c.Address, ";")
	for _, addr := range addrs {
		conn, err := net.Dial(c.Network, addr)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "net.Dial error:%v\n", err)
			continue
			//return err
		}

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			_ = tcpConn.SetKeepAlive(true)
		}
		c.innerWriter = conn
		return nil
	}
	return fmt.Errorf("hava no valid logs service addr:%v", c.Address)
}

func (c *Conn) needToConnectOnMsg() bool {
	if c.Reconnect {
		c.Reconnect = false
		return true
	}

	if c.innerWriter == nil {
		return true
	}

	if c.illNetFlag {
		return true
	}
	return false
	//return c.Options.ReconnectOnMsg
}

func (c *Conn) println(msg *Message) (err error) {
	c.Lock()
	defer c.Unlock()
	var txt string
	if c.Format != nil {
		txt = c.Format(msg)
	} else {
		txt = msg.Content
	}
	if msg.Level >= LevelError {
		txt = txt + "\n" + msg.Stack
	}
	_, err = c.innerWriter.Write(append([]byte(txt), '\n'))
	return err
}
