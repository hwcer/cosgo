package session

import "errors"

var (
	ErrorStorageNotSet    = errors.New("session storage not set")
	ErrorSessionIdEmpty   = errors.New("session id empty")
	ErrorSessionLocked    = errors.New("session locked")
	ErrorSessionTypeError = errors.New("session type error")
	ErrorSessionNotExist  = errors.New("session not exist")
)
