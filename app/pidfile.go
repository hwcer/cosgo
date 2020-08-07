package app

import (
	"cosgo/logger"
	"fmt"
	"github.com/coreos/etcd/pkg/fileutil"
	"os"
	"strconv"
	"strings"
	"github.com/shirou/gopsutil/process"
)

func init()  {
	logger.Debug("app pidfile init")
}



func initPidFile() error  {
	return writePidFile()
}



func writePidFile() (err error) {
	pidFile := Flag.GetString("pidfile")
	if pidFile == ""{
		return nil
	}
	var pid int
	err, pid = checkPidFile(pidFile)
	if err != nil {
		return err
	}
	if pid != 0 {
		exist, err := process.PidExists(int32(pid))
		if err != nil {
			return err
		}
		if exist{
			return fmt.Errorf("process %v exist, check it", pid)
		} else {
			err = deletePidFile()
			if err != nil {
				return err
			}
		}
	}

	var fhdl *os.File
	fhdl, err = os.OpenFile(pidFile, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer fhdl.Close()
	_, err = fhdl.WriteString(fmt.Sprintf("%v", os.Getpid()))
	return err
}



func deletePidFile() error {
	pidFile := Flag.GetString("pidfile")
	if pidFile == ""{
		return nil
	}
	return os.Remove(pidFile)
}


func checkPidFile(pidFile string) (error, int) {
	if !fileutil.Exist(pidFile){
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



