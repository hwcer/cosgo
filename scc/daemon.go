// Deprecated: 请改用 package await。

package scc

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync"
	"sync/atomic"
)

// Daemon 守护协程,在协程退出时自动重启协程，一般只用在主服务器中

func NewDaemon(scc *SCC) *Daemon {
	d := &Daemon{scc: scc}
	d.workers = make(chan *Worker, 10)
	return d
}

type Worker struct {
	Handle handle
	// Cancel 由worker goroutine写入、Stop调用读取,访问须经setCancel/callCancel(内部持锁),
	// 直接读写该字段无法保证并发安全
	Cancel  context.CancelFunc
	mu      sync.Mutex
	stopped atomic.Int32 // atomic: 0 未停,1 已停
}

func (w *Worker) Stop() {
	w.stopped.Store(1)
	w.callCancel()
}

// setCancel 写入Cancel(仅worker goroutine调用)
func (w *Worker) setCancel(cancel context.CancelFunc) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.Cancel = cancel
}

// callCancel 调用当前Cancel(任意goroutine)
func (w *Worker) callCancel() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.Cancel != nil {
		w.Cancel()
	}
}

func (w *Worker) isStopped() bool {
	return w.stopped.Load() != 0
}

type Daemon struct {
	scc     *SCC
	workers chan *Worker
	started atomic.Int32
}

// Start 守护协程,协程异常退出时会自动重启协程,一般使用在随主进程启动的固定协程
func (d *Daemon) Start(f handle) *Worker {
	if d.started.CompareAndSwap(0, 1) {
		d.monitor()
	}
	w := &Worker{Handle: f}
	d.workers <- w
	return w
}

func (d *Daemon) handle(w *Worker) {
	// Add(1) 必须在 go 之前,避免 Wait 在 goroutine Add 前就看到计数 0
	d.scc.WaitGroup.Go(func() {
		defer func() {
			if e := recover(); e != nil {
				d.scc.Catch(fmt.Errorf("%v\n%v", e, string(debug.Stack())))
			}
			w.callCancel()
			// worker 未被显式 Stop 且 scc 未退出,则重新入队重启;
			// 否则直接丢弃,避免向已无消费者的 channel 发送而阻塞
			if !w.isStopped() && !d.scc.Stopped() {
				select {
				case d.workers <- w:
				default:
					// 队列已满,记录并放弃重启
					d.scc.Catch(fmt.Errorf("daemon: workers queue full, drop worker"))
				}
			}
		}()
		var ctx context.Context
		var cancel context.CancelFunc
		ctx, cancel = d.scc.WithCancel()
		w.setCancel(cancel)
		w.Handle(ctx)
	})
}

func (d *Daemon) monitor() {
	d.scc.CGO(func(ctx context.Context) {
		for !d.scc.Stopped() {
			select {
			case <-ctx.Done():
				return
			case w := <-d.workers:
				d.handle(w)
			}
		}
	})
}
