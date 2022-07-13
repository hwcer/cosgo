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

func NewFileAdapter(filename string) *FileAdapter {
	f := &FileAdapter{
		Level:      LevelDebug,
		Daily:      true,
		MaxDays:    30,
		Append:     true,
		PermitMask: "0777",
		MaxLines:   0,
		MaxSize:    10 * 1024 * 1024,
		Filename:   filename,
	}
	return f
}

type FileAdapter struct {
	sync.RWMutex
	fileWriter           *os.File
	maxSizeCurSize       int
	maxLinesCurLines     int
	dailyOpenDate        int
	dailyOpenTime        time.Time
	fileNameOnly, suffix string

	Level      int                   `json:"level"`
	Append     bool                  `json:"append"`
	MaxLines   int                   `json:"maxlines"`
	MaxSize    int                   `json:"maxsize"`
	Daily      bool                  `json:"daily"`
	MaxDays    int64                 `json:"maxdays"`
	PermitMask string                `json:"permit"`
	Filename   string                `json:"filename"`
	Format     func(*Message) string `json:"_"`
}

func (f *FileAdapter) Init() error {
	if f.Level < 0 || f.Level > len(levelPrefix) {
		return errorLevelInvalid
	}
	if len(f.Filename) == 0 {
		return errors.New("must have filename")
	}
	f.suffix = filepath.Ext(f.Filename)
	f.fileNameOnly = strings.TrimSuffix(f.Filename, f.suffix)
	f.MaxSize *= 1024 * 1024 // 将单位转换成MB
	if f.suffix == "" {
		f.suffix = ".log"
	}
	return f.newFile()
}

func (f *FileAdapter) needCreateFresh(size int, day int) bool {
	return (f.MaxLines > 0 && f.maxLinesCurLines >= f.MaxLines) ||
		(f.MaxSize > 0 && f.maxSizeCurSize+size >= f.MaxSize) ||
		(f.Daily && day != f.dailyOpenDate)
}

// WriteMsg write adapter message into file.
func (f *FileAdapter) Write(msg *Message, level int) error {
	if level < f.Level {
		return nil
	}
	var txt string
	if f.Format != nil {
		txt = f.Format(msg)
	} else {
		txt = msg.String()
	}
	if level >= LevelError {
		txt = txt + "\n" + msg.Stack
	}

	day := msg.Time.Day()
	txt += "\n"
	if f.Append {
		f.RLock()
		if f.needCreateFresh(len(txt), day) {
			f.RUnlock()
			f.Lock()
			if f.needCreateFresh(len(txt), day) {
				if err := f.createFreshFile(msg.Time); err != nil {
					fmt.Fprintf(os.Stderr, "createFreshFile(%q): %s\n", f.Filename, err)
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
	perm, err := strconv.ParseInt(f.PermitMask, 8, 64)
	if err != nil {
		return nil, err
	}
	fd, err := os.OpenFile(f.Filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
	if err == nil {
		// Make sure file perm is user set perm cause of `os.OpenFile` will obey umask
		os.Chmod(f.Filename, os.FileMode(perm))
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
	fd, err := os.Open(f.Filename)
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
	rotatePerm, err := strconv.ParseInt(f.PermitMask, 8, 64)
	if err != nil {
		return err
	}

	_, err = os.Lstat(f.Filename)
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
		return fmt.Errorf("Cannot find free log number to rename %s", f.Filename)
	}
	f.fileWriter.Close()

	// 当创建新文件标记为true时
	// 当日志文件超过最大限制行
	// 当日志文件超过最大限制字节
	// 当日志文件隔天更新标记为true时
	// 将旧文件重命名，然后创建新文件
	err = os.Rename(f.Filename, fName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "os.Rename %s to %s err:%s\n", f.Filename, fName, err.Error())
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
	dir := filepath.Dir(f.Filename)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) (returnErr error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "Unable to delete old log '%s', error: %v\n", path, r)
			}
		}()

		if info == nil {
			return
		}

		if f.MaxDays != -1 && !info.IsDir() && info.ModTime().Add(24*time.Hour*time.Duration(f.MaxDays)).Before(time.Now()) {
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
