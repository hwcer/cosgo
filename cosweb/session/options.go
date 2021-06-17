package session

import (
	"context"
	"net/http"
)

var Options = struct {
	Name      string //session cookie name
	MaxAge    int64  //有效期(S)
	Secret    string //16位秘钥
	Storage   Storage
	Heartbeat int32 //心跳(S)
}{
	Name:      "_cosweb_cookie_name",
	MaxAge:    3600,
	Heartbeat: 10,
}

type Context interface {
	GetCookie(key string) (*http.Cookie, error)
	SetCookie(cookie *http.Cookie)
}

type Storage interface {
	TTL(sid string) (int64, error)                     //SESSION到期时间
	Get(sid string) (map[string]interface{}, error)    //获取session镜像数据
	Set(sid string, data map[string]interface{}) error //设置(覆盖)session数据
	Lock(sid string) bool
	UnLock(sid string) bool
	Create(map[string]interface{}) string //用户登录创建新session
	Delete(string) bool                   //退出登录删除SESSION
	Expire(sid string) error              //更新有效期
	Start(ctx context.Context) error      //启动服务器时初始化SESSION Storage
	Close() error                         //关闭服务器时断开连接等
}

func Start(ctx context.Context) error {
	if Options.Storage == nil {
		Options.Storage = NewMemory()
	}
	return Options.Storage.Start(ctx)
}
func Close() error {
	if Options.Storage != nil {
		return Options.Storage.Close()
	} else {
		return nil
	}
}

/*
func INCR(source, val interface{}) (interface{}, error) {
	switch source.(type) {
	case int:
		if v, o := val.(int); o {
			return source.(int) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case int8:
		if v, o := val.(int8); o {
			return source.(int8) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case int16:
		if v, o := val.(int16); o {
			return source.(int16) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case int32:
		if v, o := val.(int32); o {
			return source.(int32) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case int64:
		if v, o := val.(int64); o {
			return source.(int64) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case uint:
		if v, o := val.(uint); o {
			return source.(uint) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case uintptr:
		if v, o := val.(uintptr); o {
			return source.(uintptr) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case uint8:
		if v, o := val.(uint8); o {
			return source.(uint8) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case uint16:
		if v, o := val.(uint16); o {
			return source.(uint16) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case uint32:
		if v, o := val.(uint32); o {
			return source.(uint32) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case uint64:
		if v, o := val.(uint64); o {
			return source.(uint64) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case float32:
		if v, o := val.(float32); o {
			return source.(float32) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	case float64:
		if v, o := val.(float64); o {
			return source.(float64) + v, nil
		} else {
			return 0, ErrorIncrValTypeError
		}
	default:
		return 0, ErrorDataTypeNotNumber
	}
}
*/
