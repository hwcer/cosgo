package mongo

import "errors"

var (
	ErrorClientNotExist  = errors.New("mongo client not exist")
	ErrorWriteModelEmpty = errors.New("WriteModel Empty")
)
