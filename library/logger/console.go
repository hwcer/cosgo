package logger

import (
	"fmt"
	"runtime"
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
	newBrush("1;41"), // Emergency          红色底
	newBrush("1;35"), // Alert              紫色
	newBrush("1;34"), // Critical           蓝色
	newBrush("1;31"), // errMsg              红色
	newBrush("1;33"), // Warn               黄色
	newBrush("1;36"), // Informational      天蓝色
	newBrush("1;32"), // Debug              绿色
	newBrush("1;32"), // Trace              绿色
}

func NewConsoleOptions() *ConsoleOptions {
	c := &ConsoleOptions{
		Colorful: runtime.GOOS != "windows",
	}
	c.Level = "DEBUG"
	return c
}

func NewConsoleAdapter(opts *ConsoleOptions) (*ConsoleAdapter, error) {
	c := &ConsoleAdapter{}
	if err := c.init(opts); err != nil {
		return nil, err
	}
	return c, nil
}

type ConsoleOptions struct {
	Options
	Colorful bool
}

type ConsoleAdapter struct {
	//sync.Mutex
	level   int
	Options *ConsoleOptions
}

func (c *ConsoleAdapter) init(opts *ConsoleOptions) (err error) {
	c.Options = opts
	if l, ok := LevelMap[c.Options.Level]; ok {
		c.level = l
	} else {
		return fmt.Errorf("无效的日志等级:%v", c.Options.Level)
	}
	return
}

func (c *ConsoleAdapter) Write(msg *Message, level int) error {
	if level < c.level {
		return nil
	}
	var txt string
	if c.Options.Format != nil {
		txt = c.Options.Format(msg)
	} else {
		txt = msg.String()
	}
	if c.Options.Colorful {
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
	//c.Lock()
	//defer c.Unlock()
	//os.Stdout.Write(append([]byte(msg), '\n'))
	fmt.Printf(msg + "\n")
}
