package logger

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

func NewNetAdapter(network, address string) *NetAdapter {
	return &NetAdapter{Network: network, Address: address, Level: LevelDebug}
}

type NetAdapter struct {
	sync.Mutex
	Level       int    `json:"level"`
	Network     string `json:"network"`
	Address     string `json:"address"`
	Reconnect   bool   `json:"reconnect"`
	Format      func(*Message) string
	innerWriter io.WriteCloser
	illNetFlag  bool //网络异常标记
}

func (c *NetAdapter) Init() error {
	if c.Level < 0 || c.Level > len(levelPrefix) {
		return errorLevelInvalid
	}

	if c.innerWriter != nil {
		c.innerWriter.Close()
		c.innerWriter = nil
	}
	return nil
}

func (c *NetAdapter) Write(msg *Message, level int) (err error) {
	if level < c.Level {
		return nil
	}

	//msg, ok := msgText.(*Message)
	//if !ok {
	//	return
	//}

	if c.needToConnectOnMsg() {
		err = c.connect()
		if err != nil {
			return
		}
		c.illNetFlag = false
	}

	//网络异常时，消息发出
	if !c.illNetFlag {
		err = c.println(msg, level)
		//网络异常，通知处理网络的go程自动重连
		if err != nil {
			c.illNetFlag = true
		}
	}

	return
}

func (c *NetAdapter) Close() {
	if c.innerWriter != nil {
		_ = c.innerWriter.Close()
	}
}

func (c *NetAdapter) connect() error {
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

func (c *NetAdapter) needToConnectOnMsg() bool {
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

func (c *NetAdapter) println(msg *Message, level int) (err error) {
	c.Lock()
	defer c.Unlock()
	var txt string
	if c.Format != nil {
		txt = c.Format(msg)
	} else {
		txt = msg.Content
	}
	if level >= LevelError {
		txt = txt + "\n" + msg.Stack
	}
	_, err = c.innerWriter.Write(append([]byte(txt), '\n'))
	return err
}
