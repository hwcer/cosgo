package session

import (
	"github.com/hwcer/cosgo/session/memory"
	"github.com/hwcer/cosgo/session/options"
)

var storage Storage
var Options = &options.Options

type Storage interface {
	Get(key string, lock bool) (uid string, data map[string]interface{}, err error)                       //获取session镜像数据
	Save(key string, data map[string]interface{}, expire int64, unlock bool) error                        //设置(覆盖)session数据
	Create(uid string, data map[string]interface{}, expire int64, lock bool) (sid, key string, err error) //用户登录创建新session
	Delete(key string) error                                                                              //退出登录删除SESSION
	Start() error                                                                                         //启动服务器时初始化SESSION Storage
	Close() error                                                                                         //关闭服务器时断开连接等
}

func Set(s Storage) {
	storage = s
}

func Get() Storage {
	return storage
}

func Start() error {
	if storage == nil {
		storage = memory.New()
	}
	return storage.Start()
}
func Close() error {
	if storage != nil {
		return storage.Close()
	} else {
		return nil
	}
}
