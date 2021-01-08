package logger

//配置文件
type config struct {
	File       *fileLogger    `json:"File,omitempty"`
	Conn       *connLogger    `json:"Conn,omitempty"`
	Console    *consoleLogger `json:"Console,omitempty"`
	TimeFormat string         `json:"TimeFormat"`
}
