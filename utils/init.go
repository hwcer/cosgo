package utils

import (
	"fmt"
	"runtime"
)

var EventWriteChanSize = 5000
var WorkerWriteChanSize = 5000

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func Try(f func(), handler ...func(interface{})) {
	defer func() {
		if err := recover(); err != nil {
			if len(handler) == 0 {
				fmt.Printf("%v", err)
			} else {
				handler[0](err)
			}
		}
	}()
	f()
}
