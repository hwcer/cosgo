package utils

import (
	"context"
	"fmt"
	"sync/atomic"
)

//Worker 工作进程管理器，并发安全
type Worker struct {
	name   string
	index  int32
	cwrite chan interface{}
	handle func(interface{})
}

//工作进程，多任务分发
type workerThread struct {
	id     int32
	stop   int32
	name   string
	cwrite chan interface{}
	handle func(interface{})
}

func NewWorker(name string, num int32, handle func(interface{})) *Worker {
	work := &Worker{
		name:   name,
		cwrite: make(chan interface{}, WorkerWriteChanSize),
		handle: handle,
	}

	for i := int32(1); i <= num; i++ {
		work.Fork()
	}
	return work
}

func (this *workerThread) start(ctx context.Context) {
	fmt.Printf("CREATE WORKER %v[%v]\n", this.name, this.id)
	for this.stop == 0 && !IsStop() {
		select {
		case <-ctx.Done():
			this.close()
		case msg := <-this.cwrite:
			if msg == nil {
				this.close()
			} else {
				this.handle(msg)
			}
		}
	}
	fmt.Printf("CLOSE WORKER %v[%v]\n", this.name, this.id)
}

//关闭房间
func (this *workerThread) close() {
	if !atomic.CompareAndSwapInt32(&this.stop, 0, 1) {
		fmt.Printf("workerThread Stop error\n")
	}
}

func (this *Worker) Emit(msg interface{}) (ret bool) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("workerThread[%v].emit error:%v\n", this.name, err)
			ret = false
		}
	}()

	select {
	case this.cwrite <- msg:
	default:
		fmt.Printf("workerThread[%v] channel full and discard:%v\n", this.name, msg)
	}
	return true
}

//创建WORKER协程
func (this *Worker) Fork() {
	id := atomic.AddInt32(&this.index, 1)
	work := &workerThread{id: id, name: this.name, cwrite: this.cwrite, handle: this.handle}
	Go(work.start)
}

//关闭Worker
func (this *Worker) Close() {
	close(this.cwrite)
}
