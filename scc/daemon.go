package scc

import (
	"context"
	"fmt"
	"runtime/debug"
	"sync/atomic"
)

// Daemon 守护协程,在协程退出时自动重启协程，一般只用在主服务器中

func NewDaemon(scc *SCC) *Daemon {
	d := &Daemon{scc: scc}
	d.workers = make(chan *Worker, 10)
	return d
}

type Worker struct {
	Handle  handle
	Cancel  context.CancelFunc
	stopped int32 // atomic: 0 未停,1 已停
}

func (w *Worker) Stop() {
	atomic.StoreInt32(&w.stopped, 1)
	if w.Cancel != nil {
		w.Cancel()
	}
}

func (w *Worker) isStopped() bool {
	return atomic.LoadInt32(&w.stopped) != 0
}

type Daemon struct {
	scc     *SCC
	workers chan *Worker
	started int32
}

// Start 守护协程,协程异常退出时会自动重启协程,一般使用在随主进程启动的固定协程
func (d *Daemon) Start(f handle) *Worker {
	if atomic.CompareAndSwapInt32(&d.started, 0, 1) {
		d.monitor()
	}
	w := &Worker{Handle: f}
	d.workers <- w
	return w
}

func (d *Daemon) handle(w *Worker) {
	// Add(1) 必须在 go 之前,避免 Wait 在 goroutine Add 前就看到计数 0
	d.scc.WaitGroup.Add(1)
	go func() {
		defer d.scc.WaitGroup.Done()
		defer func() {
			if e := recover(); e != nil {
				d.scc.Catch(fmt.Errorf("%v\n%v", e, string(debug.Stack())))
			}
			if w.Cancel != nil {
				w.Cancel()
			}
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
		ctx, w.Cancel = d.scc.WithCancel()
		w.Handle(ctx)
	}()
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
