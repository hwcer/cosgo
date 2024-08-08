package cosgo

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	AppDir                      string = "dir" //工作目录
	AppName                     string = "name"
	AppPprof                    string = "pprof"
	AppDebug                    string = "debug"
	AppConfigNamePidFile        string = "pid"           //pid file dir
	AppConfigNameLogsPath       string = "logs.path"     //logs dir
	AppConfigNameLogsLevel      string = "logs.level"    //logs level
	AppConfigNameConfigFileExt  string = "ConfigFileExt" //config file default ext name,只能在程序中设置
	AppConfigNameConfigFileName string = "config"        //config file

)

//var (
//	//debug   bool
//	appDir  string
//	appName string
//	//workDir string
//)

func init() {
	Config.Flags(AppName, "", "", "app name")
	Config.Flags(AppPprof, "", "", "pprof server address")
	Config.Flags(AppDebug, "", false, "developer model")
	Config.Flags(AppDir, "", "", "working directory")
	Config.Flags(AppConfigNamePidFile, "", "", "pid file")
	//Config.Flags(AppConfigNameLogsPath, "", "", "logs dir")
	Config.Flags(AppConfigNameConfigFileName, "c", "", "use config file")
	Config.SetDefault(AppConfigNameConfigFileExt, "toml")
	//var (
	//	tmpDir string
	//)
	workDir, _ := os.Getwd()
	appBinFile, _ := exec.LookPath(os.Args[0])
	tmpDir, appName := filepath.Split(appBinFile)

	if filepath.IsAbs(appBinFile) {
		appDir := filepath.Dir(tmpDir)
		workDir = filepath.Dir(appDir)
	} else {
		appDir := filepath.Join(workDir, filepath.Dir(appBinFile))
		workDir, _ = filepath.Split(appDir)
		workDir = filepath.Dir(workDir)
	}

	if ext := filepath.Ext(appName); ext != "" {
		appName = strings.TrimSuffix(appName, ext)
	}
	Config.SetDefault(AppName, appName)
	Config.SetDefault(AppDir, workDir)
}
