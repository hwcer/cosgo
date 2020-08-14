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

func buildInit() error {
	args := os.Args
	if nil == args || len(args) < 2 {
		return nil
	}
	if "-v" == args[1] {
		fmt.Printf("BuildName:%v\n", BuildName)
		fmt.Printf("BuildTime:%v\n",BuildTime)
		fmt.Printf("BuildVersion:%v\n", BuildVersion)
		os.Exit(0)
	}else if "-h" == args[1]{
		pflag.Usage()
		os.Exit(0)
	}
	return nil
}
