package cosgo

import "github.com/hwcer/logger"

func init() {
	logger.SetCallDepth(4)
	logger.SetPathTrim("cosgo")
}
