package await

import (
	"sync/atomic"
)

func NewInitialize() *Initialize {
	i := new(Initialize)
	i.initializing = make(chan struct{})
	return i
}

// Initialize 并发模式下初始化控制器
type Initialize struct {
	status       int32 //0-未完成初始，1-初始化中，2-完成初始化
	initializing chan struct{}
}

func (i *Initialize) Try(handle func() error) error {
	if i.status > 1 {
		return nil
	}
	if !atomic.CompareAndSwapInt32(&i.status, 0, 1) {
		<-i.initializing
		return nil
	}
	defer func() {
		i.status = 2
		close(i.initializing)
	}()
	return handle()
}

//func (i *Initialize) Reset() {
//	if atomic.CompareAndSwapInt32(&i.status, 2, 3) {
//		i.initializing = make(chan struct{})
//		i.status = 0
//	}
//}
