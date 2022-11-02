package cosgo

import (
	"github.com/hwcer/cosgo/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
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
var SCC = utils.NewSCC(nil)

func init() {
	Config.Flags("name", "", "", "app name")
	Config.Flags("pprof", "", "", "pprof server address")
	Config.Flags("debug", "", false, "developer model")
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

	ext := filepath.Ext(appBinFile)
	if ext != "" {
		appName = strings.TrimSuffix(appName, ext)
	}
}
