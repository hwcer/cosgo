package cosgo

import (
	"github.com/hwcer/logger"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 其他模块可以使用pflag设置额外的参数
// Config 值读取优先级
// 1. overrides  use Config.Set()
// 2. flags
// 3. env. variables
// 4. config file
// 5. key/value store
// 6. defaults

var Config = &config{Viper: viper.New()}

type config struct {
	*viper.Viper
}

func (this *config) Flags(name, shorthand string, value interface{}, usage string) interface{} {
	switch v := value.(type) {
	case string:
		return pflag.StringP(name, shorthand, v, usage)
	case bool:
		return pflag.BoolP(name, shorthand, v, usage)
	case int:
		return pflag.IntP(name, shorthand, v, usage)
	case int8:
		return pflag.Int8P(name, shorthand, v, usage)
	case int16:
		return pflag.Int16P(name, shorthand, v, usage)
	case int32:
		return pflag.Int32P(name, shorthand, v, usage)
	case int64:
		return pflag.Int64P(name, shorthand, v, usage)
	case uint:
		return pflag.UintP(name, shorthand, v, usage)
	case uint8:
		return pflag.Uint8P(name, shorthand, v, usage)
	case uint16:
		return pflag.Uint16P(name, shorthand, v, usage)
	case uint32:
		return pflag.Uint32P(name, shorthand, v, usage)
	case uint64:
		return pflag.Uint64P(name, shorthand, v, usage)
	case float32:
		return pflag.Float32P(name, shorthand, v, usage)
	case float64:
		return pflag.Float64P(name, shorthand, v, usage)
	default:
		logger.Fatal("unknown type")
	}
	return nil
}

func (this *config) init() (err error) {
	pflag.Parse()
	if err = this.BindPFlags(pflag.CommandLine); err != nil {
		return err
	}
	ext := Config.GetString(AppConfigNameConfigFileExt)
	//通过基础配置读取  config.toml
	if configFile := this.GetString(AppConfigNameConfigFileName); configFile != "" {
		if s := filepath.Ext(configFile); s == "" {
			configFile = strings.Join([]string{configFile, ext}, ".")
		}
		f := Abs(configFile)
		this.SetConfigFile(f)
		if err = this.ReadInConfig(); err != nil {
			return err
		}
	}
	//通过程序名称读取 ${appName}.toml
	if f, exist := fileExist(appName, ext); exist {
		this.SetConfigFile(f)
		if err = this.MergeInConfig(); err != nil {
			return err
		}
	}
	debug = this.GetBool(AppConfigDebug)
	if debug {
		logger.SetLevel(logger.LevelDebug)
	} else {
		logger.SetLevel(logger.LevelTrace)
	}
	//设置pidfile
	if pidfile := this.GetString(AppConfigNamePidFile); pidfile != "" {
		file := Abs(pidfile)
		stat, osErr := os.Stat(file)
		if osErr != nil && !os.IsExist(osErr) {
			return osErr
		}
		if stat.IsDir() {
			file = filepath.Join(file, appName+".pid")
		}
		this.Set(AppConfigNamePidFile, file)
	}
	//设置日志
	if logsdir := this.GetString(AppConfigNameLogsDir); logsdir != "" {
		logsdir = Abs(logsdir)
		_, osErr := os.Stat(logsdir)
		if osErr != nil && !os.IsExist(osErr) {
			if err = os.MkdirAll(logsdir, 0777); err != nil {
				return
			}
		}
		this.Set(AppConfigNameLogsDir, logsdir)
		logsFile := logger.NewFile(logsdir)
		logsFile.SetFileName(logsFileNameFormatter)
		if err = logger.SetOutput("file", logsFile); err != nil {
			return
		}
	}
	return
}

func fileExist(name, ext string) (f string, exist bool) {
	var file string
	if strings.Contains(name, ".") {
		file = name
	} else {
		file = strings.Join([]string{appName, ext}, ".")
	}
	f = Abs(file)
	stat, err := os.Stat(f)
	if err == nil && !stat.IsDir() {
		exist = true
	}
	return
}

func logsFileNameFormatter() (name string, expire time.Duration) {
	name, expire = logger.DefaultFileNameFormatter()
	name = strings.Join([]string{appName, name, "log"}, ".")
	return
}
