package app

import (
	"context"
	"sync"
)

// 模块接口
type Module interface {
	ID() string
	Ready() error
	Start(context.Context, *sync.WaitGroup) error
	Stop() error
}