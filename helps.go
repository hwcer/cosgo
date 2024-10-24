package cosgo

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
	"runtime"
)

var Version string = "unknown"

func init() {
	Config.Flags("help", "h", false, "Show App helps")
	Config.Flags("version", "v", false, "Show build Branch")
}

func helps() error {
	if Config.IsSet("version") {
		fmt.Printf("runtime : %v\n", runtime.Version())
		fmt.Printf("version : %v\n", Version)
		os.Exit(0)
	} else if Config.IsSet("help") {
		pflag.Usage()
		os.Exit(0)
	}
	return nil
}
