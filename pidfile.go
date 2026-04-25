package cosgo

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/shirou/gopsutil/process"
)

var enablePidFile = false

func writePidFile() (err error) {
	file := Config.GetString(AppConfigNamePidFile)
	if file == "" {
		return
	}
	var pid int
	pid, err = checkPidFile(file)
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

	// 原子写入: 先写临时文件,再 rename 替换旧文件。
	// 避免进程在 WriteString 中途 crash 导致残留半写内容,也避免未 O_TRUNC 造成的尾部残留。
	content := []byte(strconv.Itoa(os.Getpid()))
	tmp, err := os.CreateTemp(filepath.Dir(file), filepath.Base(file)+".*.tmp")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	if _, err = tmp.Write(content); err != nil {
		tmp.Close()
		_ = os.Remove(tmpName)
		return err
	}
	if err = tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	if err = os.Rename(tmpName, file); err != nil {
		_ = os.Remove(tmpName)
		return err
	}
	enablePidFile = true
	return nil
}

func deletePidFile() error {
	if !enablePidFile {
		return nil
	}
	file := Config.GetString(AppConfigNamePidFile)
	return os.Remove(file)
}

func checkPidFile(pidFile string) (int, error) {
	if !Exist(pidFile) {
		return 0, nil
	}
	fhdl, err := os.Open(pidFile)
	if err != nil {
		return 0, err
	}
	defer fhdl.Close()
	buf := make([]byte, 64)
	n, err := fhdl.Read(buf)
	if err != nil {
		return 0, err
	}
	str := strings.TrimSpace(string(buf[:n]))
	pid, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return pid, nil
}
