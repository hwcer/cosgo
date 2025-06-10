package session

import (
	"github.com/hwcer/cosgo/session/storage"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/logger"
)

func NewSetter(id string, data any) storage.Setter {
	d := &Setter{}
	switch v := data.(type) {
	case *Data:
		d.Data = v
	case Data:
		d.Data = &v
	case map[string]any:
		d.Data = NewData("", v)
	case values.Values:
		d.Data = NewData(id, v)
	default:
		d.Data = NewData("", nil)
		logger.Alert("NewSetter Data Type Error:%v", data)
	}
	d.Data.id = id
	d.KeepAlive()
	return d
}

type Setter struct {
	*Data //数据接口
}

func (this *Setter) Get() interface{} {
	return this.Values
}

func (this *Setter) Set(data interface{}) {
	var v map[string]any
	switch i := data.(type) {
	case map[string]any:
		v = i
	case values.Values:
		v = i
	default:
		logger.Alert("Setter Set Args Data Type Error:%v", data)
	}
	if v != nil {
		this.Data.Update(v)
	}
}
