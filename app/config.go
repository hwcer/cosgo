package app

import (
	"github.com/hwcer/cosgo/library/logger"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

//其他模块可以使用pflag设置额外的参数
//Config 值读取优先级
// 1. overrides  use Config.Set()
// 2. flags
// 3. env. variables
// 4. config file
// 5. key/value store
// 6. defaults
var (
	debug   bool
	appDir  string
	appName string
	workDir string
)
var Config *viper.Viper

const (
	AppConfigNamePidFile    string = "pidfile"
	AppConfigNameLogsDir    string = "logsdir"
	AppConfigNameConfigFile string = "config"
)

func init() {
	Config = viper.New()
	pflag.String(AppConfigNamePidFile, "", "app pid file")
	pflag.String(AppConfigNameLogsDir, "", "app logs dir")
	pflag.String("name", "", "app name")
	pflag.String("pprof", "", "pprof server address")
	pflag.Bool("debug", false, "developer model")
	pflag.StringP("config", "c", "", "use config file")

	var (
		tmpDir     string
		appBinFile string
	)
	workDir, _ = os.Getwd()
	appBinFile, _ = exec.LookPath(os.Args[0])
	tmpDir, appName = filepath.Split(appBinFile)

	if filepath.IsAbs(appBinFile) {
		appDir = filepath.Dir(tmpDir)
		workDir = filepath.Dir(appDir)
	} else {
		appDir = filepath.Join(workDir, filepath.Dir(appBinFile))
		workDir, _ = filepath.Split(appDir)
		workDir = filepath.Dir(workDir)
	}

	ext := filepath.Ext(appBinFile)
	if ext != "" {
		appName = strings.TrimSuffix(appName, ext)
	}
}

func loggerMessageFormat(message *logger.Message) string {
	return "[" + message.Level + "] " + message.Content
}

func initFlag() (err error) {
	pflag.Parse()
	Config.BindPFlags(pflag.CommandLine)
	//通过配置读取
	if configFile := Config.GetString(AppConfigNameConfigFile); configFile != "" {
		f := Abs(configFile)
		//Config.AddConfigPath(f)
		Config.SetConfigFile(f)
		err = Config.ReadInConfig()
		if err != nil {
			return err
		}
	}
	debug = Config.GetBool("debug")
	//设置pidfile
	if pidfile := Config.GetString(AppConfigNamePidFile); pidfile != "" {
		file := Abs(pidfile)
		stat, osErr := os.Stat(file)
		if osErr != nil && !os.IsExist(osErr) {
			return osErr
		}
		if stat.IsDir() {
			file = filepath.Join(file, appName+".pid")
		}
		Config.Set(AppConfigNamePidFile, file)
	}
	//设置日志
	if logsdir := Config.GetString(AppConfigNameLogsDir); logsdir != "" {
		logsdir = Abs(logsdir)
		_, osErr := os.Stat(logsdir)
		if osErr != nil && !os.IsExist(osErr) {
			if err = os.MkdirAll(logsdir, 0777); err != nil {
				return
			}
		}
		Config.Set(AppConfigNameLogsDir, logsdir)
		logsFile := filepath.Join(logsdir, appName+".log")
		opts := logger.NewFileOptions()
		opts.Filename = Abs(logsFile)
		if Config.GetBool("Debug") {
			opts.Level = "DEBUG"
		} else {
			opts.Level = "INFO"
		}
		if loggerFileAdapter, err = logger.NewFileAdapter(opts); err != nil {
			return
		}
	}
	return nil
}

//Abs 获取以工作路径为起点的绝对路径
func Abs(dir ...string) string {
	path := filepath.Join(dir...)
	if !filepath.IsAbs(path) {
		path = filepath.Join(workDir, path)
	}
	return path
}

func Debug() bool {
	return debug
}

//Dir APP主程序所在目录
func Dir() string {
	return appDir
}

//Name 项目内部获取appName
func Name() string {
	return appName
}

//WorkDir 程序工作目录
func WorkDir() string {
	return workDir
}
