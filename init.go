package cosgo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	AppConfigDebug              string = "debug"
	AppConfigNamePidFile        string = "pid"           //pid file dir
	AppConfigNameLogsDir        string = "logs"          //logs dir
	AppConfigNameDaemonize      string = "daemonize"     //后台运行
	AppConfigNameConfigFileExt  string = "ConfigFileExt" //config file default ext name,只能在程序中设置
	AppConfigNameConfigFileName string = "config"        //config file

)

var (
	debug   bool
	appDir  string
	appName string
	workDir string
)

func init() {
	Config.Flags(AppConfigNamePidFile, "", "", "pid file")
	Config.Flags(AppConfigNameLogsDir, "", "", "logs dir")
	Config.Flags("name", "", "", "app name")
	Config.Flags("pprof", "", "", "pprof server address")
	Config.Flags(AppConfigDebug, "", false, "developer model")
	Config.Flags(AppConfigNameDaemonize, "", false, "后台运行和使用nohup一个效果")
	Config.Flags(AppConfigNameConfigFileName, "c", "", "use config file")
	Config.SetDefault(AppConfigNameConfigFileExt, "toml")
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

	if ext := filepath.Ext(appName); ext != "" {
		appName = strings.TrimSuffix(appName, ext)
	}
}
