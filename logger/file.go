package logger

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

type fileNameFormatter func() (name string, expire time.Duration)

// DefaultFileNameFormatter 默认日志文件,每日一份
func DefaultFileNameFormatter() (name string, expire time.Duration) {
	t := time.Now()
	r := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location()).AddDate(0, 0, 1)
	name = t.Format("20060102")
	expire = time.Duration(r.Unix()-t.Unix()) * time.Second
	return
}

func NewFile(path string) *File {
	return &File{logsPath: path}
}

type File struct {
	file     *os.File
	fileName fileNameFormatter     //日志名规则
	logsPath string                //日志目录
	Sprintf  func(*Message) string //格式化message
}

func (this *File) GetFileName() (name string, expire time.Duration) {
	if this.fileName != nil {
		return this.fileName()
	}
	return DefaultFileNameFormatter()
}

// SetFileName 设置日志文件名,  前缀(string) 或者 fileNameFormatter
func (this *File) SetFileName(f fileNameFormatter) {
	this.fileName = f
}

func (this *File) Init() (err error) {
	f, err := os.Stat(this.logsPath)
	if err != nil {
		return err
	}
	if !f.IsDir() {
		return fmt.Errorf("path not dir:%v", this.logsPath)
	}
	if err = this.mayCreateFile(); err != nil {
		return
	}
	return
}

func (this *File) Write(msg *Message) (err error) {
	if this.file == nil {
		return errors.New("file handle empty")
	}
	b := bytes.Buffer{}
	if this.Sprintf != nil {
		b.WriteString(this.Sprintf(msg))
	} else {
		b.WriteString(msg.String())
	}
	b.WriteString("\n")
	if msg.Level >= LevelError {
		b.WriteString(msg.Stack)
		b.WriteString("\n")
	}
	_, err = b.WriteTo(this.file)
	//_, err = this.file.Write([]byte(txt))
	return
}

func (this *File) timer() {
	_ = this.mayCreateFile()
}

func (this *File) mayCreateFile() (err error) {
	// Open the log file
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("%v", e)
		}
	}()
	name, expire := this.GetFileName()
	defer func() {
		time.AfterFunc(expire, this.timer)
	}()

	if this.file != nil && name == filepath.Base(this.file.Name()) {
		return nil
	}
	var perm int64
	perm, err = strconv.ParseInt("0777", 8, 64)
	if err != nil {
		return
	}
	path := filepath.Join(this.logsPath, name)
	fd, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, os.FileMode(perm))
	if err != nil {
		return err
	}
	fd.Fd()
	_ = os.Chmod(path, os.FileMode(perm))
	if err = fd.Sync(); err != nil {
		return
	}
	var old *os.File
	if old, this.file = this.file, fd; old != nil {
		time.AfterFunc(5*time.Second, func() {
			_ = old.Close()
		})
	}
	return
}

func (this *File) pathExists(path string) error {
	f, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !f.IsDir() {
		return fmt.Errorf("path not dir:%v", path)
	}
	return nil
}
