package logger

import (
	"strings"
	"time"
)

type Message struct {
	Path    string
	Time    time.Time
	Level   Level
	Stack   string
	Content string
}

func (this *Message) String() string {
	b := strings.Builder{}
	b.WriteString(this.Time.Format(DefTimeFormat))
	b.WriteString(" [")
	b.WriteString(this.Level.String())
	b.WriteString("] ")
	if this.Path != "" {
		b.WriteString("[")
		b.WriteString(this.Path)
		b.WriteString("] ")
	}
	b.WriteString(this.Content)
	return b.String()
}
