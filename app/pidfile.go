package app

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"github.com/shirou/gopsutil/process"
)

func writePidFile() error {
	var err error
	err, pid := CheckPidFile(a.pidFile)
	if err != nil {
		return err
	}
	if pid != 0 {
		exist, err := IsProcessExist(pid)
		if err != nil {
			return err
		}
		if exist == true {
			return fmt.Errorf("process %v exist, check it", pid)
		} else {
			err = DeletePidFile(a.pidFile)
			if err != nil {
				return err
			}
		}
	}

	err = sysutil.WritePidFile(a.pidFile)
	if err != nil {
		return err
	}

	return nil
}

func  deletePidFile() {
	DeletePidFile(a.pidFile)
}



func IsFileExist(filename string) bool {
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		return false
	}
	return true
}


func CheckPidFile(filedir string) (error, int) {
	if IsFileExist(filedir) == true {

		fhdl, err := os.Open(filedir)
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
	} else {
		return nil, 0
	}
}

func WritePidFile(filedir string) error {

	fhdl, err := os.OpenFile(filedir, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}
	defer fhdl.Close()

	_, err = fhdl.WriteString(fmt.Sprintf("%v", os.Getpid()))

	return err
}

func DeletePidFile(filedir string) error {
	return os.Remove(filedir)
}
func IsProcessExist(pid int) (bool, error) {
	return process.PidExists(int32(pid))
}
