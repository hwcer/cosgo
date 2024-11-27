package logger

import (
	"fmt"
)

// Output Output输出时是否对字体染色
type Output interface {
	Init() error
	Write(message *Message) (err error)
}

func (this *Logger) SetOutput(name string, output Output) error {
	if _, ok := this.outputs[name]; ok {
		return fmt.Errorf("adapter name exist:%v", name)
	}
	if err := output.Init(); err != nil {
		return err
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	dict := make(map[string]Output)
	for k, v := range this.outputs {
		dict[k] = v
	}
	dict[name] = output
	this.outputs = dict
	return nil
}

func (this *Logger) DelOutput(name string) {
	if _, ok := this.outputs[name]; !ok {
		return
	}
	this.mutex.Lock()
	defer this.mutex.Unlock()
	dict := make(map[string]Output)
	for k, v := range this.outputs {
		if k != name {
			dict[k] = v
		}
	}
	this.outputs = dict
}
