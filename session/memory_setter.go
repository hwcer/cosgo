package session

import (
	"github.com/hwcer/cosgo/storage"
	"time"
)

func NewSetter(id storage.MID, data interface{}) storage.Setter {
	d := &Setter{
		Data: *storage.NewData(id, data),
		//locked: 1,
	}
	if Options.MaxAge > 0 {
		d.Expire(Options.MaxAge)
	}
	return d
}

type Setter struct {
	storage.Data       //数据接口
	expire       int64 //过期时间
}

// Expire 设置有效期(s)
func (this *Setter) Expire(s int64) {
	this.expire = time.Now().Unix() + s
}
