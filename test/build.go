package main

import (
	"fmt"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
)

//其他模块可以使用pflag设置额外的参数
var Flag *viper.Viper

func init() {
	Flag = viper.New()
	//解析所有命令行参数
	pflag.Bool("debug", false, "developer model")
	pflag.String("name", "", "应用名称")
	pflag.BoolP("help", "h", false, "显示帮助信息")
	pflag.BoolP("version", "v", false, "显示版本号")
	//设置默认值
	Flag.SetDefault("name", "就当我是傻瓜好了")
}

func parse() error {
	pflag.Parse()
	Flag.BindPFlags(pflag.CommandLine)
	if Flag.IsSet("version") {
		fmt.Printf("我也不知道版本号")
		os.Exit(0)
	} else if Flag.IsSet("help") {
		pflag.Usage()
		os.Exit(0)
	}
	return nil
}
