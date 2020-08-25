package app

import (
	"fmt"
	"github.com/spf13/pflag"
	"os"
)


var (
	BuildName    string
	BuildTime    string
	BuildVersion string
)

func init()  {
	pflag.BoolP("help","h",false,"Show App helps")
	pflag.BoolP("version","v",false,"Show Build Version")
}

func buildInit() error {
	if Flag.IsSet("version") {
		fmt.Printf("BuildName:%v\n", BuildName)
		fmt.Printf("BuildTime:%v\n",BuildTime)
		fmt.Printf("BuildVersion:%v\n", BuildVersion)
		os.Exit(0)
	}else if Flag.IsSet("help") {
		pflag.Usage()
		os.Exit(0)
	}
	return nil
}
