package network

import (
	"runtime"
	"sync"
)


var stop int32 //停止标志



var msgqueId uint32 //消息队列id
var msgqueMap = map[uint32]IMsgQue{}
var msgqueMapSync sync.Mutex

var Config = struct {
	AutoCompressLen uint32
	UdpServerGoCnt  int
	PoolSize        int32
	SSLCrtPath      string
	SSLKeyPath      string
	EnableWss       bool
	ReadDataBuffer  int
	StopTimeout     int
}{UdpServerGoCnt: 64, PoolSize: 50000, ReadDataBuffer: 1 << 12, StopTimeout: 3000}

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	if Logger == nil {
		SetLogger(&defaultLogger{})
	}
}
