package apps

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
var Config *viper.Viper

func init() {
	Config = viper.New()
	pflag.Bool("debug", false, "developer model")
	pflag.String("logdir", "", "app logs dir")
	pflag.String("pidfile", "", "app pid file")
	pflag.String("profile", "", "profile address")
	pflag.StringP("config", "c", "", "use config file")

	var (
		tmpDir      string
		appName     string
		appBinDir   string
		appWorkDir  string
		appExecFile string
	)

	appWorkDir, _ = os.Getwd()
	appExecFile, _ = exec.LookPath(os.Args[0])
	tmpDir, appName = filepath.Split(appExecFile)

	if filepath.IsAbs(appExecFile) {
		appBinDir = filepath.Dir(tmpDir)
		appWorkDir = filepath.Dir(appBinDir)
	} else {
		appBinDir = filepath.Join(appWorkDir, filepath.Dir(appExecFile))
		appWorkDir, _ = filepath.Split(appBinDir)
		appWorkDir = filepath.Dir(appWorkDir)
	}

	ext := filepath.Ext(appExecFile)
	if ext != "" {
		appName = strings.TrimRight(appName, ext)
	}

	Config.SetDefault("name", appName)
	Config.SetDefault("logdir", filepath.Join(appWorkDir, "logs"))
	Config.SetDefault("pidfile", filepath.Join(appBinDir, appName+".pid"))
	Config.SetDefault("appBinDir", appBinDir)
	Config.SetDefault("appWorkDir", appWorkDir)
	Config.SetDefault("appExecFile", appExecFile)

}

func initFlag() error {
	pflag.Parse()
	Config.BindPFlags(pflag.CommandLine)
	//通过配置读取
	if Config.IsSet("config") {
		f := Config.GetString("config")
		appWorkDir := Config.GetString("appWorkDir")
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
	logdir := Config.GetString("logdir")
	if logdir != "" {
		logger.SetLogPathTrim(Config.GetString(logdir))
	}
	return nil
}
