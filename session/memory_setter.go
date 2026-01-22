package session

import (
	"github.com/hwcer/cosgo/storage"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/logger"
)

func NewMemorySetter(id string, data any) storage.Setter {
	var d *Data
	switch v := data.(type) {
	case *Data:
		d = v
	case Data:
		d = &v
	case map[string]any:
		d = NewData(id, v)
	case values.Values:
		d = NewData(id, v)
	default:
		d = NewData(id, nil)
		logger.Alert("NewMemorySetter Data Type Error:%v", data)
	}
	d.id = id //重置成Setter id
	d.KeepAlive()
	return d
}

//type MemorySetter struct {
//	*Data //数据接口
//}
//
//func (this *MemorySetter) Get() interface{} {
//	return this.Values
//}
//
//func (this *MemorySetter) Set(data interface{}) {
//	var v map[string]any
//	switch i := data.(type) {
//	case map[string]any:
//		v = i
//	case values.Values:
//		v = i
//	default:
//		logger.Alert("MemorySetter Set Args Data Type Error:%v", data)
//	}
//	if v != nil {
//		this.Data.Update(v)
//	}
//}
