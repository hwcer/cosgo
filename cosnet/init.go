package cosnet

import (
	"cosgo/utils"
	"runtime"
	"strings"
	"time"
)

var scc *utils.SCC
var servers []Server
var timestamp int

var Config = struct {
	Heartbeat        int   //(MS)服务器心跳,用来检测玩家僵尸连接
	AutoCompressSize int32 //自动压缩
	UdpServerGoCnt   int

	SSLCrtPath     string
	SSLKeyPath     string
	EnableWss      bool
	ReadDataBuffer int
	StopTimeout    int

	WriteChanSize  int32 //写通道缓存
	ConnectMaxSize int32 //连接人数
	ConnectTimeout int   //连接超时(MS),至少大于ServerInterval2倍以上

	MsgDataType MsgDataType //默认包体编码方式
}{
	Heartbeat:      2000,
	UdpServerGoCnt: 64,
	ReadDataBuffer: 1 << 12,
	StopTimeout:    3000,

	WriteChanSize:  500,
	ConnectMaxSize: 50000,
	ConnectTimeout: 10000,

	MsgDataType: MsgDataTypeProto,
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	scc = utils.NewSCC()
}
func startHeartbeat() {
	scc.CGO(func(stop chan struct{}) {
		t := time.Millisecond * time.Duration(Config.Heartbeat)
		ticker := time.NewTimer(t)
		defer ticker.Stop()
		for !scc.Stopped() {
			select {
			case <-stop:
				return
			case <-ticker.C:
				timestamp += Config.Heartbeat
				ticker.Reset(t)
			}
		}
	})
}

//启动服务器,根据
func NewServer(addr string, handler Handler) (srv Server) {
	addrs := strings.Split(addr, "://")
	if addrs[0] == "tcp" {
		srv = NewTcpServer(addrs[1], handler)
	} else if addrs[0] == "udp" {
		//TODO UDP
	} else if addrs[0] == "ws" || addrs[0] == "wss" {
		//TODO wss
	}
	servers = append(servers, srv)
	return
}

func SCC() *utils.SCC {
	return scc
}

func Start() (err error) {
	startHeartbeat()
	for _, srv := range servers {
		if err = srv.start(); err != nil {
			return err
		}
	}
	return nil
}

func Close() (err error) {
	//for _, srv := range servers {
	//	srv.start()
	//}
	return scc.Close()
}
func Stopped() bool {
	return scc.Stopped()
}
