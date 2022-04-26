package cosnet

import (
	"context"
	"errors"
	"github.com/hwcer/cosgo/storage/cache"
	"github.com/hwcer/cosgo/utils"
	"net/url"
	"strings"
	"time"
)

type Server interface {
	Start()
	Close() error
}

func newSocketSetter(id uint64, val interface{}) cache.Dataset {
	dataset := val.(cache.Dataset)
	dataset.Reset(id, nil)
	return dataset
}

func New(ctx context.Context) *Agents {
	agents := &Agents{
		scc:     utils.NewSCC(ctx),
		emitter: NewEmitter(8),
		Cache:   *cache.New(1024),
		Handler: NewHandle(),
	}
	agents.Cache.NewSetter = newSocketSetter
	agents.scc.CGO(agents.heartbeat)
	return agents
}

//Agents socket管理器
type Agents struct {
	cache.Cache
	scc     *utils.SCC
	emitter *Emitter //事件触发器
	servers []Server //全局关闭时需要关闭的服务
	Handler Handler  //消息处理器
}

func (this *Agents) create(socket *Socket) {
	this.Cache.Push(socket)
	this.emitter.Emit(EventTypeConnected, socket)
}

//remove 移除Socket
func (this *Agents) remove(socket *Socket) {
	if s := this.Cache.Delete(socket.Id()); s != nil {
		this.emitter.Emit(EventTypeDisconnect, socket)
	}
}

//heartbeat 启动协程定时清理无效用户
func (this *Agents) heartbeat(ctx context.Context) {
	t := time.Millisecond * time.Duration(Options.SocketHeartbeat)
	ticker := time.NewTimer(t)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			utils.Try(this.doHeartbeat)
			ticker.Reset(t)
		}
	}
}
func (this *Agents) doHeartbeat() {
	this.Cache.Range(func(v cache.Dataset) bool {
		socket := v.(*Socket)
		socket.Heartbeat()
		this.emitter.Emit(EventTypeHeartbeat, socket)
		return true
	})
}

func (this *Agents) Close(timeout time.Duration) error {
	if !this.scc.Cancel() {
		return nil
	}
	for _, s := range this.servers {
		s.Close()
	}
	this.servers = nil
	return this.scc.Wait(timeout)
}

//New 创建新socket并自动加入到Agents管理器
func (this *Agents) New(io NetIO, netType NetType) (*Socket, error) {
	if !netType.Any(AnyNetType) || !netType.Any(AnyNetwork) {
		return nil, errors.New("netType error")
	}
	socket := &Socket{
		io:      io,
		agents:  this,
		cwrite:  make(chan Message, Options.WriteChanSize),
		netType: netType,
		Data:    *cache.NewData(),
	}
	socket.start()
	this.create(socket)
	return socket, nil
}

//On 注册事件监听器
func (this *Agents) On(e EventType, f EventsFunc) {
	this.emitter.On(e, f)
}

//Socket 通过SOCKETID获取SOCKET
func (this *Agents) Socket(id uint64) (*Socket, bool) {
	if v, ok := this.Cache.Get(id); !ok {
		return nil, false
	} else if v2, ok2 := v.(*Socket); ok2 {
		return v2, true
	} else {
		return nil, false
	}
}

//Broadcast 广播,filter 过滤函数，如果不为nil且返回false则不对当期socket进行发送消息
func (this *Agents) Broadcast(msg Message, filter func(*Socket) bool) {
	this.Cache.Range(func(v cache.Dataset) bool {
		sock := v.(*Socket)
		if sock.io != nil && (filter == nil || filter(sock)) {
			sock.Write(msg)
		}
		return true
	})
}

//Listen 启动柜服务器,监听address
func (this *Agents) Listen(address string) (server Server, err error) {
	var addrs *url.URL
	addrs, err = url.Parse(address)
	if err != nil {
		return nil, err
	}
	//b, _ := json.Marshal(addrs)
	//fmt.Printf("%v\n", string(b))
	if addrs.Scheme == "tcp" {
		server, err = NewTcpServer(this, addrs)
	} else if addrs.Scheme == "udp" {
		server, err = NewUdpServer(this, addrs)
	} else if addrs.Scheme == "ws" || addrs.Scheme == "wss" {
		server, err = NewWssServer(this, addrs)
	} else {
		err = errors.New("address proto error")
	}
	if err == nil && server != nil {
		this.servers = append(this.servers, server)
		this.scc.GO(server.Start)
	}
	return
}

//Connect 连接服务器address
func (this *Agents) Connect(address string) (socket *Socket, err error) {
	addrs := strings.Split(address, "://")
	if len(addrs) != 2 {
		return nil, errors.New("Request Address Errror")
	}
	network := strings.ToLower(addrs[0])
	if network == "tcp" {
		socket, err = NewTcpConnect(this, addrs[1])
	} else if addrs[0] == "udp" {

	} else if addrs[0] == "ws" || addrs[0] == "wss" {

	} else {
		err = errors.New("address proto error")
	}

	return
}
