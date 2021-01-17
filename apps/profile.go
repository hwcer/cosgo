package apps

import (
	"cosgo/debug"
	"runtime"
)

//性能调优

func initProfile() error {
	runtime.GOMAXPROCS(runtime.NumCPU())
	profile := Config.GetString("profile")
	if profile != "" {
		debug.StartPprofSrv(profile)
	}

	return nil
}
