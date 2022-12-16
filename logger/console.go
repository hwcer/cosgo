package logger

import (
	"math/rand"
	"os"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type brush func(string) string

func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

// 鉴于终端的通常使用习惯，一般白色和黑色字体是不可行的,所以30,37不可用，
var colors = []brush{
	newBrush("1;32"), // Trace              绿色
	newBrush("1;32"), // Debug              绿色
	newBrush("1;36"), // Info              天蓝色
	newBrush("1;33"), // Warn               黄色
	newBrush("1;31"), // Error              红色
	newBrush("1;34"), // Alert              蓝色
	newBrush("1;35"), // FATAL              紫色
	newBrush("1;41"), // Emergency          红色底
}

func NewConsoleAdapter() *ConsoleAdapter {
	t := time.Now().UnixNano()
	name := strconv.FormatInt(t, 36) + "-" + strconv.FormatInt(rand.Int63n(89999)+10000, 36)
	return &ConsoleAdapter{Colorful: true, name: name}
}

type ConsoleAdapter struct {
	sync.Mutex
	name     string
	Format   func(*Message) string
	Colorful bool
}

func (c *ConsoleAdapter) Name() string {
	return c.name
}
func (c *ConsoleAdapter) Init() (err error) {
	if runtime.GOOS == "windows" {
		c.Colorful = false
	}
	return
}

func (c *ConsoleAdapter) Write(msg *Message) error {
	var txt string
	level := msg.Level
	if c.Format != nil {
		txt = c.Format(msg)
	} else {
		txt = msg.String()
	}
	if c.Colorful {
		txt = colors[int(level)](txt)
	}
	if level >= LevelError {
		txt = txt + "\n" + msg.Stack
	}
	return c.printlnConsole(txt)
}

func (c *ConsoleAdapter) Close() {

}

func (c *ConsoleAdapter) printlnConsole(msg string) (err error) {
	c.Lock()
	defer c.Unlock()
	_, err = os.Stdout.Write(append([]byte(msg), '\n'))
	return
}
