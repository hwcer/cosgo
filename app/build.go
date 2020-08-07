package app


import (
	"fmt"
	"os"
)

// 程序编译信息 参数有编译脚本传入
//DINF_VER  = "unknown" // 版本号
//DINF_SRC  = "unknown" // 源码分支及版本信息
//DINF_BTM  = "unknown" // 编译时间
//DINF_GO   = "unknown" // golang版本
//DINF_HOST = "unknown" // host info

var (
	BuildVersion string
	BuildTime    string
	BuildName    string
)

func init() {
	args := os.Args
	if nil == args || len(args) < 2 {
		return
	}
	if "-v" == args[1] {
		fmt.Printf("%s: v%s (%s)\n", BuildName, BuildVersion, BuildTime)
	} else if "-h" == args[1] {
		fmt.Println("Usage:")
		fmt.Printf("./%s\n", BuildName)
		fmt.Printf("./%s -v\n", BuildName)
		fmt.Printf("./%s -h\n", BuildName)
	}
	os.Exit(0)
}
