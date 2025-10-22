package session

import (
	"github.com/hwcer/cosgo/session/storage"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/logger"
)

func NewMemorySetter(id string, data any) storage.Setter {
	d := &MemorySetter{}
	switch v := data.(type) {
	case *Data:
		d.Data = v
	case Data:
		d.Data = &v
	case map[string]any:
		d.Data = NewData(id, v)
	case values.Values:
		d.Data = NewData(id, v)
	default:
		d.Data = NewData(id, nil)
		logger.Alert("NewMemorySetter Data Type Error:%v", data)
	}
	d.Data.id = id //重置成Setter id
	d.KeepAlive()
	return d
}

type MemorySetter struct {
	*Data //数据接口
}

func (this *MemorySetter) Get() interface{} {
	return this.Values
}

func (this *MemorySetter) Set(data interface{}) {
	var v map[string]any
	switch i := data.(type) {
	case map[string]any:
		v = i
	case values.Values:
		v = i
	default:
		logger.Alert("MemorySetter Set Args Data Type Error:%v", data)
	}
	if v != nil {
		this.Data.Update(v)
	}
}
