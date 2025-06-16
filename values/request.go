package values

import (
	"encoding/json"
	"errors"
)

// 仅仅用于发送请求，且服务器使用 Message响应

type Request struct {
	Code int32 `json:"code"`
	Data Bytes `json:"data"`
}

func (this *Request) Error() (r string) {
	if this.Code == 0 {
		r = "not Error"
	} else {
		if err := json.Unmarshal(this.Data, &r); err != nil {
			r = err.Error()
		}
	}
	return
}

func (this *Request) Marshal(v any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	this.Data = b
	return nil
}

// Unmarshal 如果是一个错误信息 应当单独处理
func (this *Request) Unmarshal(i interface{}) (err error) {
	if this.Code != 0 {
		var s string
		if err = json.Unmarshal(this.Data, &s); err != nil {
			return
		} else {
			return errors.New(s)
		}
	}
	if len(this.Data) > 0 {
		err = json.Unmarshal(this.Data, i)
	}
	return
}
