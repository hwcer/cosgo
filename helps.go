package cosgo

import (
	"github.com/spf13/pflag"
	"os"
)

//var (
//	BUIND_GO   = "unknown"
//	BUIND_VER  = "unknown"
//	BUIND_TIME = "unknown"
//)

func init() {
	Config.Flags("help", "h", false, "Show App helps")
	//Config.Flags("version", "v", false, "Show build Version")
}

func helps() error {
	if Config.IsSet("help") {
		pflag.Usage()
		os.Exit(0)
	}
	return nil
}
