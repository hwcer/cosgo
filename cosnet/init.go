package cosnet

import (
	"github.com/hwcer/cosgo/cosnet/message"
	"runtime"
	"strings"
)

var Config = struct {
	AutoCompressSize int32 //自动压缩
	UdpServerGoCnt   int

	SSLCrtPath     string
	SSLKeyPath     string
	EnableWss      bool
	ReadDataBuffer int
	StopTimeout    int

	WriteChanSize  int32 //写通道缓存
	ConnectMaxSize int32 //连接人数

	ReconnectTime   int //断线重连等待心跳次数，0-不会断线重连
	SocketTimeout   int //连接超时几次心跳没有动作被判断掉线
	SocketHeartbeat int //(MS)服务器心跳,用来检测玩家僵尸连接

	MsgDataType message.ContentType //默认包体编码方式
}{

	UdpServerGoCnt: 64,
	ReadDataBuffer: 1 << 12,
	StopTimeout:    3000,

	WriteChanSize:  500,
	ConnectMaxSize: 50000,

	ReconnectTime:   15,
	SocketTimeout:   5,
	SocketHeartbeat: 2000,

	MsgDataType: message.ContentTypeProto,
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

//启动服务器,根据
func NewServer(addr string, handler Handler) (srv Server) {
	addrs := strings.Split(addr, "://")
	if addrs[0] == "tcp" {
		return NewTcpServer(addrs[1], handler)
	} else if addrs[0] == "udp" {
		//TODO UDP
	} else if addrs[0] == "ws" || addrs[0] == "wss" {
		//TODO wss
	}
	return nil
}
