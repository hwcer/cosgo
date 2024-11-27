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
	stopped bool
}

func (w *Worker) Stop() {
	w.stopped = true
	if w.Cancel != nil {
		w.Cancel()
	}
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
	go func() {
		defer func() {
			if e := recover(); e != nil {
				d.scc.Catch(fmt.Errorf("%v\n%v", e, string(debug.Stack())))
			}
			w.Cancel()
			if !w.stopped {
				d.workers <- w
			}
		}()
		d.scc.WaitGroup.Add(1)
		defer d.scc.WaitGroup.Done()
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
