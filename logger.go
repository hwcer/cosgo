package cosgo

import (
	"github.com/hwcer/cosgo/logger"
	"strings"
)

func init() {
	logger.SetCallDepth(0)
	if v := logger.GetDefaultAdapter(); v != nil {
		v.Format = loggerMessageFormat
	}
	logger.SetLogPathTrim(workDir)
}

func loggerMessageFormat(msg *logger.Message) string {
	b := strings.Builder{}
	b.WriteString(msg.Time.Format(Options.DataTimeFormat))
	b.WriteString(" [")
	b.WriteString(msg.Level.String())
	b.WriteString("] ")
	b.WriteString(msg.Content)
	return b.String()
}
