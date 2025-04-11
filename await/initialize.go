package await

import (
	"sync/atomic"
)

func NewInitialize() *Initialize {
	i := new(Initialize)
	i.c = make(chan struct{})
	return i
}

// Initialize 并发模式下初始化控制器
type Initialize struct {
	n int32 //0-未完成初始，1-初始化中，2-完成初始化
	c chan struct{}
}

func (i *Initialize) Try(f func() error) (err error) {
	n, c := i.n, i.c
	if n > 1 {
		return
	}
	if !atomic.CompareAndSwapInt32(&i.n, 0, 1) {
		<-c
		return
	}
	defer func() {
		i.n = 2
		i.c = make(chan struct{})
		close(c)
		i.n = 0
	}()
	return f()
}
