package app

import (
	"bytes"
	"encoding/json"
	goarg "github.com/alexflint/go-arg"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
)


var (
	appLogDir    string
	appPidDir    string
	appBinDir    string
	appExecDir   string
	appWorkDir   string
	appDataDir   string
	appConfigDir   string
)

var flags  = struct {
	Debug 				 bool    `arg:"-d, help:app run in develop mod"`        //DEBUG模式
	Shard                string  `arg:"-s, help:shard name"`                    //指定shard
	Config               string  `arg:"-c, help:app config path"`              //逻辑配置
	Version              bool    `arg:"-v, help:show version info and exit"`   // 显示版本信息
}{

}


func init() {

	goarg.MustParse(&flags)


	appWorkDir, _ = os.Getwd()
	tmpDir := ""
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

	appConfDir = filepath.Join(appWorkDir, "conf")
	appLogDir = filepath.Join(appWorkDir, "log")
	appOssLogDir = filepath.Join(appWorkDir, "logoss")
	appPidDir = filepath.Join(appWorkDir, "pid")
	appDataDir = filepath.Join(appWorkDir, "data")

	// try load from file
	envFile := filepath.Join(appConfDir, "env.json")
	_, err := os.Stat(envFile)
	if err == nil || os.IsExist(err) {
		jsonFile, err := os.Open(envFile)
		defer jsonFile.Close()
		if err == nil {
			ec := &envConf{}
			byteValue, _ := ioutil.ReadAll(jsonFile)
			byteValue = bytes.TrimPrefix(byteValue, []byte("\xef\xbb\xbf"))
			err = json.Unmarshal(byteValue, ec)
			if err == nil {
				env = ec.Env
				appShard = ec.Shard
				project = ec.Project
				confdAddrs = ec.Confd
				rpcdEtcds = ec.Etcd
				consulAddrs = ec.Consul
				appPlatAddr = ec.PlatAddr
			} else {
				log.Printf("load env decode failed:%v", err)
			}
		} else {
			log.Printf("load env.json failed:%v", err)
		}
	}


	// confd rpcd
	//appConfdPrefix = path.Join("/"+appProject, appEnv, "confd")
	appRpcdPrefix = path.Join("/"+appProject, appEnv, "rpcd")
	appStatdPrefix = path.Join("/"+appProject, appEnv, "statd")

	log.Printf(">> APP ENV: %v, CONFD:%v\n", env, confdAddrs)
}
