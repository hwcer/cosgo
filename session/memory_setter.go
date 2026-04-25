package session

import (
	"github.com/hwcer/cosgo/storage"
	"github.com/hwcer/cosgo/values"
	"github.com/hwcer/logger"
)

// NewMemorySetter 内存后端的 Setter 工厂
// 将传入的 data 适配为 *Data 并重置 id 为 storage 分配的 token
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
	d.id = id
	d.KeepAlive()
	return d
}
