package logger

import (
	"os"
	"runtime"
	"sync"
)

type brush func(string) string

func newBrush(color string) brush {
	pre := "\033["
	reset := "\033[0m"
	return func(text string) string {
		return pre + color + "m" + text + reset
	}
}

//鉴于终端的通常使用习惯，一般白色和黑色字体是不可行的,所以30,37不可用，
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
	return &ConsoleAdapter{Colorful: true, Level: LevelDebug}
}

type ConsoleAdapter struct {
	sync.Mutex
	Level    int
	Format   func(*Message) string
	Colorful bool
}

func (c *ConsoleAdapter) Init() (err error) {
	if c.Level < 0 || c.Level > len(levelPrefix) {
		return errorLevelInvalid
	}
	if runtime.GOOS == "windows" {
		c.Colorful = false
	}
	return
}

func (c *ConsoleAdapter) Write(msg *Message, level int) error {
	if level < c.Level {
		return nil
	}
	var txt string
	if c.Format != nil {
		txt = c.Format(msg)
	} else {
		txt = msg.String()
	}
	if c.Colorful {
		txt = colors[level](txt)
	}
	if level >= LevelError {
		txt = txt + "\n" + msg.Stack
	}
	c.printlnConsole(txt)
	return nil
}

func (c *ConsoleAdapter) Close() {

}

func (c *ConsoleAdapter) printlnConsole(msg string) {
	c.Lock()
	defer c.Unlock()
	os.Stdout.Write(append([]byte(msg), '\n'))
}
