package app

import (
	"fmt"
	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/shirou/gopsutil/process"
	"os"
	"strconv"
	"strings"
)

var enablePidFile = false

func writePidFile() (err error) {
	file := Config.GetString(AppConfigNamePidFile)
	if file == "" {
		return
	}
	var pid int
	err, pid = checkPidFile(file)
	if err != nil {
		return err
	}
	if pid != 0 {
		var exist bool
		exist, err = process.PidExists(int32(pid))
		if err != nil {
			return err
		}
		if exist {
			return fmt.Errorf("process %v exist, check it", pid)
		} else {
			err = deletePidFile()
			if err != nil {
				return err
			}
		}
	}

	var fhdl *os.File
	fhdl, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer fhdl.Close()
	enablePidFile = true
	_, err = fhdl.WriteString(fmt.Sprintf("%v", os.Getpid()))
	return err
}

func deletePidFile() error {
	if !enablePidFile {
		return nil
	}
	file := Config.GetString(AppConfigNamePidFile)
	return os.Remove(file)
}

func checkPidFile(pidFile string) (error, int) {
	if !fileutil.Exist(pidFile) {
		return nil, 0
	}
	fhdl, err := os.Open(pidFile)
	if err != nil {
		return err, 0
	}
	defer fhdl.Close()
	buf := make([]byte, 64)
	n, err := fhdl.Read(buf)
	if err != nil {
		return err, 0
	}
	str := string(buf[:n])
	str = strings.TrimSpace(str)
	pid, err := strconv.Atoi(str)
	if err != nil {
		return err, 0
	}
	return nil, pid
}
