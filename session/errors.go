package session

import (
	"github.com/hwcer/cosgo/values"
)

var (
	ErrorStorageEmpty     = values.Errorf(202, "session Storage not set")
	ErrorSessionEmpty     = values.Errorf(203, "session token empty")
	ErrorSessionNotCreate = values.Errorf(204, "session not create")
	ErrorSessionNotExist  = values.Errorf(205, "session not exist")
	ErrorSessionExpired   = values.Errorf(206, "session expire")
	ErrorSessionIllegal   = values.Errorf(207, "session illegal")
	ErrorSessionUnknown   = values.Errorf(208, "session unknown error")
	ErrorSessionReplaced  = values.Errorf(209, "session replaced")
)

func Errorf(format any, args ...any) error {
	return values.Errorf(201, format, args...)
}
