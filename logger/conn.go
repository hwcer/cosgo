package logger

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"sync"
)

func NewNetOptions() *NetOptions {
	c := &NetOptions{}
	c.Level = "DEBUG"
	return c
}

func NewNetAdapter(opts *NetOptions) (*NetAdapter, error) {
	f := &NetAdapter{}
	if err := f.init(opts); err != nil {
		return nil, err
	}
	return f, nil
}

type NetOptions struct {
	Options
	Net       string `json:"net"`
	Addr      string `json:"addr"`
	Reconnect bool   `json:"reconnect"`
}

type NetAdapter struct {
	sync.Mutex
	level       int
	Options     *NetOptions
	innerWriter io.WriteCloser
	illNetFlag  bool //网络异常标记
}

func (c *NetAdapter) init(opts *NetOptions) error {
	c.Options = opts
	if l, ok := LevelMap[c.Options.Level]; ok {
		c.level = l
	} else {
		return fmt.Errorf("无效的日志等级:%v", c.Options.Level)
	}
	if c.innerWriter != nil {
		c.innerWriter.Close()
		c.innerWriter = nil
	}
	return nil
}

func (c *NetAdapter) Write(msg *Message, level int) (err error) {
	if level < c.level {
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
		//重连成功
		c.illNetFlag = false
	}

	//每条消息都重连一次日志中心，适用于写日志频率极低的情况下的服务调用,避免长时间连接，占用资源
	//if c.Options.ReconnectOnMsg { // 频繁日志发送切勿开启
	//	defer c.innerWriter.Close()
	//}

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
		c.innerWriter.Close()
	}
}

func (c *NetAdapter) connect() error {
	if c.innerWriter != nil {
		c.innerWriter.Close()
		c.innerWriter = nil
	}
	addrs := strings.Split(c.Options.Addr, ";")
	for _, addr := range addrs {
		conn, err := net.Dial(c.Options.Net, addr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "net.Dial error:%v\n", err)
			continue
			//return err
		}

		if tcpConn, ok := conn.(*net.TCPConn); ok {
			tcpConn.SetKeepAlive(true)
		}
		c.innerWriter = conn
		return nil
	}
	return fmt.Errorf("hava no valid logs service addr:%v", c.Options.Addr)
}

func (c *NetAdapter) needToConnectOnMsg() bool {
	if c.Options.Reconnect {
		c.Options.Reconnect = false
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
	if c.Options.Format != nil {
		txt = c.Options.Format(msg)
	} else {
		txt = msg.Content
	}
	if level >= LevelError {
		txt = txt + "\n" + msg.Stack
	}
	_, err = c.innerWriter.Write(append([]byte(txt), '\n'))
	return err
}
