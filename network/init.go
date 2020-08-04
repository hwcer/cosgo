package network

import (
	"runtime"
	"sync"
	"sync/atomic"
)

type WaitGroup struct {
	count int64
}

func (r *WaitGroup) Add(delta int) {
	atomic.AddInt64(&r.count, int64(delta))
}

func (r *WaitGroup) Done() {
	atomic.AddInt64(&r.count, -1)
}

func (r *WaitGroup) Wait() {
	for atomic.LoadInt64(&r.count) > 0 {
		Sleep(1)
	}
}

func (r *WaitGroup) TryWait() bool {
	return atomic.LoadInt64(&r.count) == 0
}

var waitAll = &WaitGroup{} //等待所有goroutine

var stop int32 //停止标志

var goid uint32
var gocount int32 //goroutine数量

var msgqueId uint32 //消息队列id
var msgqueMap = map[uint32]IMsgQue{}
var msgqueMapSync sync.Mutex

var stopChanForGo = make(chan struct{})

var poolChan = make(chan func())
var poolGoCount int32

var StartTick int64
var NowTick int64
var Timestamp int64

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
	timerTick()
}
