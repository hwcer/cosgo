package cosnet

import (
	"cosgo/logger"
	"runtime"
	"strings"
	"sync"
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
	ConnectTimeout int32 //连接超时(MS),至少大于ServerInterval2倍以上

	ServerInterval int64 //(MS)服务器心跳,用来检测玩家僵尸连接

	MsgDataType MsgDataType //默认包体编码方式
}{
	UdpServerGoCnt: 64,
	ReadDataBuffer: 1 << 12,
	StopTimeout:    3000,

	WriteChanSize:  500,
	ConnectMaxSize: 50000,
	ConnectTimeout: 6000,
	ServerInterval: 2000,

	MsgDataType: MsgDataTypeProto,
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func Go(wgp *sync.WaitGroup, f func()) {
	go func() {
		wgp.Add(1)
		defer wgp.Done()
		f()
	}()
}

func Try(f func(), handler ...func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			if len(handler) == 0 {
				logger.Error("%v", err)
			} else {
				handler[0](err)
			}
		}
	}()
	f()
}

func SafeGo(wgp *sync.WaitGroup, f func(), handler ...func(interface{})) {
	Go(wgp, func() {
		Try(f, handler...)
	})
}

//启动服务器,根据
func NewServer(addr string, handler Handler) (srv Server, err error) {
	addrs := strings.Split(addr, "://")
	if addrs[0] == "tcp" {
		srv, err = NewTcpServer(addrs[1], handler)
	} else if addrs[0] == "udp" {
		//TODO UDP
	} else if addrs[0] == "ws" || addrs[0] == "wss" {
		//TODO wss
	}
	return
}
