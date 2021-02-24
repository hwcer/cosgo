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
	appName    string
	appWorkDir string
	Config     *viper.Viper
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
	appWorkDir, _ = os.Getwd()
	appBinFile, _ = exec.LookPath(os.Args[0])
	tmpDir, appName = filepath.Split(appBinFile)

	if filepath.IsAbs(appBinFile) {
		appBinDir = filepath.Dir(tmpDir)
		appWorkDir = filepath.Dir(appBinDir)
	} else {
		appBinDir = filepath.Join(appWorkDir, filepath.Dir(appBinFile))
		appWorkDir, _ = filepath.Split(appBinDir)
		appWorkDir = filepath.Dir(appWorkDir)
	}

	ext := filepath.Ext(appBinFile)
	if ext != "" {
		appName = strings.TrimRight(appName, ext)
	}
	//设置配置默认值
	Config.SetDefault("logs", filepath.Join(appWorkDir, "logs"))
	Config.SetDefault("pidfile", filepath.Join(appWorkDir, appName+".pid"))
}

func initFlag() error {
	pflag.Parse()
	Config.BindPFlags(pflag.CommandLine)
	//通过配置读取
	if Config.IsSet("Config") {
		f := Config.GetString("Config")
		if !filepath.IsAbs(f) {
			f = filepath.Join(appWorkDir, f)
		}
		Config.AddConfigPath(f)
		err := Config.ReadInConfig()
		if err != nil {
			return err
		}
	}
	//设置日志
	logs := Config.GetString("logs")
	if !filepath.IsAbs(logs) {
		logs = filepath.Join(appWorkDir, logs)
	}
	logger.SetLogPathTrim(logs)

	return nil
}

//GetAppName 项目内部获取appName
func GetAppName() string {
	return appName
}

//GetAppWorkDir 项目内部获取appWorkDir
func GetAppWorkDir() string {
	return appWorkDir
}
