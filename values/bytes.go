package values

import (
	"encoding/json"
)

type Bytes []byte

const (
	BytesEmptyJson = "null"
)

func (b *Bytes) MarshalJSON() ([]byte, error) {
	if b == nil || len(*b) == 0 {
		return []byte(BytesEmptyJson), nil
	}
	return *b, nil
}

func (b *Bytes) UnmarshalJSON(v []byte) error {
	// 输入是 JSON "null" 时保持接收器为空,否则原样存入
	if string(v) != BytesEmptyJson {
		*b = v
	}
	return nil
}

func (b *Bytes) Marshal(v any) error {
	if v == nil {
		return nil
	}
	d, err := json.Marshal(v)
	if err == nil {
		*b = d
	}
	return err
}

func (b *Bytes) Unmarshal(i any) error {
	if len(*b) == 0 {
		return nil
	}
	return json.Unmarshal(*b, i)
}
