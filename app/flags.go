package app

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
)


var (
	appName      string
	appBinDir    string
	appExecDir   string
	appWorkDir   string
)


//其他模块可以使用pflag设置额外的参数
var Flag *viper.Viper

func init() {
	Flag =  viper.New()
	pflag.Bool("debug",false,"developer model")
	pflag.String("name","cosjs","app name")
	pflag.String("logdir","","log path")
	pflag.String("pidfile","","app pid file")


	var tmpDir string
	appWorkDir, _ = os.Getwd()
	appExecDir, _ = exec.LookPath(os.Args[0])
	tmpDir, appName = filepath.Split(appExecDir)

	if filepath.IsAbs(appExecDir) {
		appBinDir = filepath.Dir(tmpDir)
		appWorkDir = filepath.Dir(appBinDir)
	} else {
		appBinDir = filepath.Join(appWorkDir, filepath.Dir(appExecDir))
		appWorkDir, _ = filepath.Split(appBinDir)
		appWorkDir = filepath.Dir(appWorkDir)
	}

}

func initFlag() error  {
	pflag.Parse()
	Flag.BindPFlags(pflag.CommandLine)
	return nil
}