package cosgo

import (
	"github.com/hwcer/cosgo/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	AppConfigNamePidFile    string = "pid"    //pid file dir
	AppConfigNameLogsDir    string = "logs"   //logs dir
	AppConfigNameConfigFile string = "config" //config file
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
	Config.Flags("config", "c", "", "use config file")

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
