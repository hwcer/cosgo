package logger

// 日志等级，从0-7，日优先级由高到低
const (
	LevelTrace = iota // 用户级基本输出
	LevelDebug        // 用户级调试
	LevelInfo         // 用户级信息
	LevelWarn         // 用户级警告
	LevelError        // 用户级错误
	LevelPANIC        //
	LevelFATAL
)

// 日志记录等级字段
var levelPrefix = []string{"TRACE", "DEBUG", "INFO ", "WARN ", "ERROR", "PANIC", "FATAL"}

// 日志等级和描述映射关系
var LevelMap = map[string]int{
	"ERROR": LevelError,
	"WARN":  LevelWarn,
	"INFO":  LevelInfo,
	"TRAC":  LevelTrace,
	"DEBUG": LevelDebug,
	"TRACE": LevelTrace,
	"FATAL": LevelFATAL,
}

// log provider interface
type adapter interface {
	Close()
	Write(msg *Message, level int) error
}

type Options struct {
	Level  string
	Format func(*Message) string
}
