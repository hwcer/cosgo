package logger

import "time"

type Message struct {
	Path    string
	Time    time.Time
	Level   string
	Stack   string
	Content string
}

func (this *Message) String() string {
	return this.Time.Format(DefTimeFormat) + " [" + this.Level + "] " + "[" + this.Path + "] " + this.Content
}
