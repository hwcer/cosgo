package await

import (
	"sync"
)

//建议直接使用 sync.Once
// 仅仅为了兼容旧版

func NewInitialize() *Initialize {
	return &Initialize{}
}

// Initialize 并发模式下初始化控制器
type Initialize struct {
	sync.Once
}

func (i *Initialize) Try(f func() error) (err error) {
	i.Once.Do(func() {
		err = f()
	})
	return
}
