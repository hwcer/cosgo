package cosnet

import (
	"runtime"
)

var Config = struct {
	AutoCompressLen uint32
	UdpServerGoCnt  int

	SSLCrtPath     string
	SSLKeyPath     string
	EnableWss      bool
	ReadDataBuffer int
	StopTimeout    int

	WriteChanSize    int32 //写通道缓存
	ConnectMaxSize   int32 //连接人数
	ConnectTimeout   int32 //连接超时(MS)，依赖于ConnectHeartbeat
	ConnectHeartbeat int32 //心跳(MS)每隔多久检查一次客户端状态

	ServerInterval int64 //服务器时钟(MS)
}{
	UdpServerGoCnt: 64,
	ReadDataBuffer: 1 << 12,
	StopTimeout:    3000,

	WriteChanSize:  500,
	ConnectMaxSize: 50000,

	ConnectTimeout:   6000,
	ConnectHeartbeat: 500,

	ServerInterval: 100,
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}
