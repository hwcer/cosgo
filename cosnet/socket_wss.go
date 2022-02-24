package cosnet

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/hwcer/cosgo/utils"
	"net"
	"net/http"
	"net/url"
)

type wssSocket struct {
	conn   *websocket.Conn
	agents *Agents
}

func (this *wssSocket) Read(head []byte) (Message, error) {
	_, data, err := this.conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	msg := this.agents.Handler.New()
	msg.Reset(data)
	return msg, nil
}

func (this *wssSocket) Write(msg Message) error {
	return this.conn.WriteMessage(websocket.BinaryMessage, msg.Data())
}

func (this *wssSocket) Close() error {
	if this.conn != nil {
		return this.conn.Close()
	}
	return nil
}

func (this *wssSocket) LocalAddr() net.Addr {
	if this.conn != nil {
		return this.conn.LocalAddr()
	}
	return nil
}

func (this *wssSocket) RemoteAddr() net.Addr {
	if this.conn != nil {
		return this.conn.RemoteAddr()
	}
	return nil
}

func NewWssServer(agent *Agents, address *url.URL) (*wssServer, error) {
	if address.Path == "" {
		address.Path = "/"
	}
	server := &wssServer{
		agents:   agent,
		address:  address,
		listener: new(http.Server),
	}
	server.upgrader = &websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	server.listener.Addr = address.Host
	server.listener.Handler = server
	return server, nil
}

type wssServer struct {
	agents   *Agents
	address  *url.URL
	listener *http.Server
	upgrader *websocket.Upgrader
}

func (this *wssServer) Close() error {
	return this.listener.Close()
}

func (s *wssServer) Start() {
	var err error
	if s.listener.TLSConfig != nil {
		err = s.listener.ListenAndServeTLS("", "")
	} else {
		err = s.listener.ListenAndServe()
	}
	if err != nil && err != utils.ErrorTimeout {
		fmt.Printf("udpServer start err:%v\n", err)
	}
}

func (this *wssServer) ServeHTTP(hw http.ResponseWriter, hr *http.Request) {
	if hr.URL.Path != this.address.Path {
		hw.WriteHeader(404)
		return
	}
	c, err := this.upgrader.Upgrade(hw, hr, nil)
	if err != nil {
		fmt.Printf("start failed msgque:%v err:%v\n", this.address.String(), err)
	} else {
		io := &wssSocket{conn: c, agents: this.agents}
		this.agents.New(io, NetworkWssServer)
	}
}
