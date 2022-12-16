package logger

type Level int8

// 日志等级，从0-7，日优先级由高到低
const (
	LevelTrace Level = iota // 用户级基本输出
	LevelDebug              // 用户级调试
	LevelInfo               // 用户级信息
	LevelWarn               // 用户级警告
	LevelError              // 用户级错误
	LevelAlert              //系统级警告，比如数据库访问异常，配置文件出错等
	LevelFATAL              //PANIC
)

// 日志记录等级字段
var levelPrefix = map[Level]string{LevelTrace: "TRACE", LevelDebug: "DEBUG", LevelInfo: "INFO", LevelWarn: "WARN", LevelError: "ERROR", LevelAlert: "ALERT", LevelFATAL: "FATAL"}

func (l Level) String() string {
	return levelPrefix[l]
}
