package app

import (
	"runtime"
)

//性能调优

func initProfile() error {
	runtime.GOMAXPROCS(runtime.NumCPU())
	return nil
}
