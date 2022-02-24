package logger

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

func NewFileOptions() *FileOptions {
	f := &FileOptions{
		Daily:      true,
		MaxDays:    30,
		Append:     true,
		PermitMask: "0777",
		MaxLines:   0,
		MaxSize:    10 * 1024 * 1024,
	}
	f.Level = "DEBUG"
	return f
}

func NewFileAdapter(opts *FileOptions) (*FileAdapter, error) {
	f := &FileAdapter{}
	if err := f.init(opts); err != nil {
		return nil, err
	}
	return f, nil
}

type FileOptions struct {
	Options
	Filename   string `json:"filename"`
	Append     bool   `json:"append"`
	MaxLines   int    `json:"maxlines"`
	MaxSize    int    `json:"maxsize"`
	Daily      bool   `json:"daily"`
	MaxDays    int64  `json:"maxdays"`
	PermitMask string `json:"permit"`
}

type FileAdapter struct {
	level   int
	Options *FileOptions
	sync.RWMutex
	fileWriter           *os.File
	maxSizeCurSize       int
	maxLinesCurLines     int
	dailyOpenDate        int
	dailyOpenTime        time.Time
	fileNameOnly, suffix string
}

func (f *FileAdapter) init(opts *FileOptions) (err error) {
	f.Options = opts
	if len(f.Options.Filename) == 0 {
		return errors.New("FileOptions must have filename")
	}
	f.suffix = filepath.Ext(f.Options.Filename)
	f.fileNameOnly = strings.TrimSuffix(f.Options.Filename, f.suffix)
	f.Options.MaxSize *= 1024 * 1024 // 将单位转换成MB
	if f.suffix == "" {
		f.suffix = ".log"
	}
	if l, ok := LevelMap[f.Options.Level]; ok {
		f.level = l
	} else {
		return fmt.Errorf("无效的日志等级:%v", f.Options.Level)
	}
	err = f.newFile()
	return err
}

func (f *FileAdapter) needCreateFresh(size int, day int) bool {
	return (f.Options.MaxLines > 0 && f.maxLinesCurLines >= f.Options.MaxLines) ||
		(f.Options.MaxSize > 0 && f.maxSizeCurSize+size >= f.Options.MaxSize) ||
		(f.Options.Daily && day != f.dailyOpenDate)
}

// WriteMsg write adapter message into file.
func (f *FileAdapter) Write(msg *Message, level int) error {
	if level < f.level {
		return nil
	}
	var txt string
	if f.Options.Format != nil {
		txt = f.Options.Format(msg)
	} else {
		txt = msg.String()
	}
	if level >= LevelError {
		txt = txt + "\n" + msg.Stack
	}

	day := msg.Time.Day()
	txt += "\n"
	if f.Options.Append {
		f.RLock()
		if f.needCreateFresh(len(txt), day) {
			f.RUnlock()
			f.Lock()
			if f.needCreateFresh(len(txt), day) {
				if err := f.createFreshFile(msg.Time); err != nil {
					fmt.Fprintf(os.Stderr, "createFreshFile(%q): %s\n", f.Options.Filename, err)
				}
			}
			f.Unlock()
		} else {
			f.RUnlock()
		}
	}

	f.Lock()
	_, err := f.fileWriter.Write([]byte(txt))
	if err == nil {
		f.maxLinesCurLines++
		f.maxSizeCurSize += len(txt)
	}
	f.Unlock()
	return err
}

func (f *FileAdapter) createLogFile() (*os.File, error) {
	// Open the log file
	perm, err := strconv.ParseInt(f.Options.PermitMask, 8, 64)
	if err != nil {
		return nil, err
	}
	fd, err := os.OpenFile(f.Options.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
	if err == nil {
		// Make sure file perm is user set perm cause of `os.OpenFile` will obey umask
		os.Chmod(f.Options.Filename, os.FileMode(perm))
	}
	return fd, err
}

func (f *FileAdapter) newFile() error {
	file, err := f.createLogFile()
	if err != nil {
		return err
	}
	if f.fileWriter != nil {
		f.fileWriter.Close()
	}
	f.fileWriter = file

	fInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("get stat err: %s", err)
	}
	f.maxSizeCurSize = int(fInfo.Size())
	f.dailyOpenTime = time.Now()
	f.dailyOpenDate = f.dailyOpenTime.Day()
	f.maxLinesCurLines = 0
	if f.maxSizeCurSize > 0 {
		count, err := f.lines()
		if err != nil {
			return err
		}
		f.maxLinesCurLines = count
	}
	return nil
}

func (f *FileAdapter) lines() (int, error) {
	fd, err := os.Open(f.Options.Filename)
	if err != nil {
		return 0, err
	}
	defer fd.Close()

	buf := make([]byte, 32768) // 32k
	count := 0
	lineSep := []byte{'\n'}

	for {
		c, err := fd.Read(buf)
		if err != nil && err != io.EOF {
			return count, err
		}

		count += bytes.Count(buf[:c], lineSep)

		if err == io.EOF {
			break
		}
	}

	return count, nil
}

// new file name like  xx.2013-01-01.001.log
func (f *FileAdapter) createFreshFile(logTime time.Time) error {
	// file exists
	// match the next available number
	num := 1
	fName := ""
	rotatePerm, err := strconv.ParseInt(f.Options.PermitMask, 8, 64)
	if err != nil {
		return err
	}

	_, err = os.Lstat(f.Options.Filename)
	if err != nil {
		// 初始日志文件不存在，无需创建新文件
		goto RESTART_LOGGER
	}
	// 日期变了， 说明跨天，重命名时需要保存为昨天的日期
	if f.dailyOpenDate != logTime.Day() {
		for ; err == nil && num <= 999; num++ {
			fName = f.fileNameOnly + fmt.Sprintf(".%s.%03d%s", f.dailyOpenTime.Format("2006-01-02"), num, f.suffix)
			_, err = os.Lstat(fName)
		}
	} else { //如果仅仅是文件大小或行数达到了限制，仅仅变更后缀序号即可
		for ; err == nil && num <= 999; num++ {
			fName = f.fileNameOnly + fmt.Sprintf(".%s.%03d%s", logTime.Format("2006-01-02"), num, f.suffix)
			_, err = os.Lstat(fName)
		}
	}

	if err == nil {
		return fmt.Errorf("Cannot find free log number to rename %s", f.Options.Filename)
	}
	f.fileWriter.Close()

	// 当创建新文件标记为true时
	// 当日志文件超过最大限制行
	// 当日志文件超过最大限制字节
	// 当日志文件隔天更新标记为true时
	// 将旧文件重命名，然后创建新文件
	err = os.Rename(f.Options.Filename, fName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.Rename %s to %s err:%s\n", f.Options.Filename, fName, err.Error())
		goto RESTART_LOGGER
	}

	err = os.Chmod(fName, os.FileMode(rotatePerm))

RESTART_LOGGER:

	startLoggerErr := f.newFile()
	go f.deleteOldLog()

	if startLoggerErr != nil {
		return fmt.Errorf("Rotate StartLogger: %s", startLoggerErr)
	}
	if err != nil {
		return fmt.Errorf("Rotate: %s", err)
	}
	return nil
}

func (f *FileAdapter) deleteOldLog() {
	dir := filepath.Dir(f.Options.Filename)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "Unable to delete old log '%s', error: %v\n", path, r)
			}
		}()

		if info == nil {
			return
		}

		if f.Options.MaxDays != -1 && !info.IsDir() && info.ModTime().Add(24*time.Hour*time.Duration(f.Options.MaxDays)).Before(time.Now()) {
			if strings.HasPrefix(filepath.Base(path), filepath.Base(f.fileNameOnly)) &&
				strings.HasSuffix(filepath.Base(path), f.suffix) {
				os.Remove(path)
			}
		}
		return
	})
}

func (f *FileAdapter) Close() {
	f.fileWriter.Close()
}
