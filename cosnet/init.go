package cosnet

import (
	"cosgo/app"
	"runtime"
	"sync"
	"time"
)

var cancel chan struct{}
var timestamp int64 //时间线(MS)
var heartbeat int64 //心跳时间(MS)

var msgqueId uint32 //消息队列id
var msgqueMap = map[uint32]Socket{}
var msgqueMapSync sync.Mutex

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
}{
	UdpServerGoCnt: 64,
	ReadDataBuffer: 1 << 12,
	StopTimeout:    3000,

	WriteChanSize:  500,
	ConnectMaxSize: 50000,

	ConnectTimeout:   6000,
	ConnectHeartbeat: 500,
}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	heartbeat = 100
	cancel = make(chan struct{})
	app.Go(timerTick)
}

func timerTick() {
	var ticker = time.NewTicker(time.Millisecond * time.Duration(heartbeat))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			timestamp += heartbeat
		case <-cancel:
			return
		}
	}
}

func Close() {
	select {
	case <-cancel:
	default:
		close(cancel)
	}
}
