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
	storage.Data //数据接口
	//uuid         string
	expire int64 //过期时间
	//locked       int32 //SESSION锁
}

//func (this *Setter) Lock() bool {
//	return atomic.CompareAndSwapInt32(&this.locked, 0, 1)
//}
//func (this *Setter) UnLock() bool {
//	return atomic.CompareAndSwapInt32(&this.locked, 1, 0)
//}

// Expire 设置有效期(s)
func (this *Setter) Expire(s int64) {
	this.expire = time.Now().Unix() + s
}
