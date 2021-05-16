package cosweb

import (
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosgo/utils"
)

type Options struct {
	SessionKey     string
	SessionType    []int //存放SESSION KEY的方式
	SessionSecret  string
	SessionStorage session.Storage //Session数据存储器
}

func NewOptions() *Options {
	return &Options{
		SessionKey:    "CosWebSessId",
		SessionType:   []int{RequestDataTypeCookie, RequestDataTypeQuery},
		SessionSecret: utils.Random.String(16),
	}
}
