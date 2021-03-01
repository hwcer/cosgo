package app

import (
	"cosgo/logger"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//其他模块可以使用pflag设置额外的参数
var (
	appDir  string
	appName string

	Config *viper.Viper
)

func init() {
	Config = viper.New()
	pflag.Bool("debug", false, "developer model")
	pflag.String("logs", "", "app logs dir")
	pflag.StringP("config", "c", "", "use config file")
	pflag.String("pidfile", "", "app pid file")
	pflag.String("profile", "", "profile address")

	var (
		tmpDir     string
		appBinDir  string
		appBinFile string
	)
	appDir, _ = os.Getwd()
	appBinFile, _ = exec.LookPath(os.Args[0])
	tmpDir, appName = filepath.Split(appBinFile)

	if filepath.IsAbs(appBinFile) {
		appBinDir = filepath.Dir(tmpDir)
		appDir = filepath.Dir(appBinDir)
	} else {
		appBinDir = filepath.Join(appDir, filepath.Dir(appBinFile))
		appDir, _ = filepath.Split(appBinDir)
		appDir = filepath.Dir(appDir)
	}

	ext := filepath.Ext(appBinFile)
	if ext != "" {
		appName = strings.TrimRight(appName, ext)
	}
	//设置配置默认值
	Config.SetDefault("logs", filepath.Join(appDir, "logs"))
	Config.SetDefault("pidfile", appName+".pid")
}

func initFlag() error {
	pflag.Parse()
	Config.BindPFlags(pflag.CommandLine)
	//通过配置读取
	if Config.IsSet("Config") {
		f := Config.GetString("Config")
		if !filepath.IsAbs(f) {
			f = filepath.Join(appDir, f)
		}
		Config.AddConfigPath(f)
		err := Config.ReadInConfig()
		if err != nil {
			return err
		}
	}
	//设置PIDfile
	pidfile := Config.GetString("pidfile")
	if !filepath.IsAbs(pidfile) {
		Config.Set("pidfile", filepath.Join(appDir, pidfile))
	}

	//设置日志
	logs := Config.GetString("logs")
	if !filepath.IsAbs(logs) {
		logs = filepath.Join(appDir, logs)
	}
	logger.SetLogPathTrim(logs)

	return nil
}

//GetDir 项目内部获取appWorkDir
func GetDir() string {
	return appDir
}

//GetName 项目内部获取appName
func GetName() string {
	return appName
}
