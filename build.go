package cosgo

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
)

var (
	BUIND_GO   = "unknown"
	BUIND_VER  = "unknown"
	BUIND_TIME = "unknown"
)

func init() {
	Config.Flags("help", "h", false, "Show App helps")
	Config.Flags("version", "v", false, "Show build Version")
}

func initBuild() error {
	if Config.IsSet("version") {
		fmt.Printf("BUIND GO:%v\n", BUIND_GO)
		fmt.Printf("BUIND VER:%v\n", BUIND_VER)
		fmt.Printf("BUIND TIME:%v\n", BUIND_TIME)
		os.Exit(0)
	} else if Config.IsSet("help") {
		pflag.Usage()
		os.Exit(0)
	}
	return nil
}
