package utils

import (
	"fmt"
	"sync"
	"sync/atomic"
)

//Worker 工作进程管理器，并发安全
type Worker struct {
	swg    sync.WaitGroup
	stop   chan struct{}
	index  int32
	cwrite chan interface{}
	handle func(interface{})
}

//工作进程，多任务分发
type workerThread struct {
	id     int32
	cwrite chan interface{}
	handle func(interface{})
}

func NewWorker(num int32, handle func(interface{})) *Worker {
	worker := &Worker{
		stop:   make(chan struct{}),
		cwrite: make(chan interface{}, WorkerWriteChanSize),
		handle: handle,
	}

	for i := int32(1); i <= num; i++ {
		worker.Fork()
	}
	return worker
}

func (this *workerThread) start(stop chan struct{}) {
	for {
		select {
		case <-stop:
			fmt.Printf("WorkerThread Stop%v\n", this.id)
			return
		case msg := <-this.cwrite:
			this.doHandle(msg)
		}
	}
}

func (this *workerThread) doHandle(msg interface{}) {
	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("panic in Go: %v\n", err)
		}
	}()
	this.handle(msg)
}

func (this *Worker) Emit(msg interface{}) (ret bool) {
	defer func() {
		if err := recover(); err != nil {
			ret = false
		}
	}()

	select {
	case this.cwrite <- msg:
	default:
		fmt.Printf("workerThread channel full and discard:%v\n", msg)
	}
	return true
}

//创建WORKER协程
func (this *Worker) Fork() {
	id := atomic.AddInt32(&this.index, 1)
	work := &workerThread{id: id, cwrite: this.cwrite, handle: this.handle}
	go func() {
		this.swg.Add(1)
		defer this.swg.Done()
		work.start(this.stop)
	}()
}

//关闭Worker
func (this *Worker) Close() {
	select {
	case <-this.stop:
	default:
		close(this.stop)
	}
	this.swg.Wait()
}
