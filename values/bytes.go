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
	if len(*b) != len(BytesEmptyJson) && string(v) != BytesEmptyJson {
		*b = v
	}
	return nil
}

func (b *Bytes) Marshal(v interface{}) error {
	if v == nil {
		return nil
	}
	d, err := json.Marshal(v)
	if err == nil {
		*b = d
	}
	return err
}

func (b *Bytes) Unmarshal(i interface{}) error {
	if len(*b) == 0 {
		return nil
	}
	return json.Unmarshal(*b, i)
}
