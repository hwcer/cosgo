package session

import (
	"github.com/hwcer/cosgo/values"
)

var (
	ErrorStorageNotSet     = values.Errorf(201, "session Storage not set")
	ErrorSessionIdEmpty    = values.Errorf(202, "session id empty")
	ErrorSessionLocked     = values.Errorf(203, "session locked")
	ErrorSessionTypeError  = values.Errorf(204, "session type error")
	ErrorSessionNotExist   = values.Errorf(205, "session not exist")
	ErrorSessionTypeExpire = values.Errorf(206, "session expire")
	ErrorSessionIllegal    = values.Errorf(207, "session illegal")
	ErrorSessionUnknown    = values.Errorf(208, "session unknown error")
	ErrorSessionReplaced   = values.Errorf(209, "session replaced")
)
