package cosgo

import "github.com/hwcer/logger"

func init() {
	logger.SetCallDepth(4)
	logger.SetPathTrim("cosgo")
	s := logger.Console.Sprintf
	On(EventTypStarted, func() error {
		logger.Console.Sprintf = nil
		return nil
	})
	On(EventTypStopping, func() error {
		logger.Console.Sprintf = s
		return nil
	})
}

// Console 控制台输出
func Console(format any, args ...any) {
	msg := &logger.Message{}
	msg.Content = logger.Sprintf(format, args...)
	_ = logger.Console.Write(msg)
}
