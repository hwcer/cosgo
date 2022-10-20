package values

import (
	"encoding/json"
)

type Bytes []byte

// TODO nil
func (b *Bytes) MarshalJSON() ([]byte, error) {
	if b == nil || len(*b) == 0 {
		return []byte("\"\""), nil
	}
	return *b, nil
}
func (b *Bytes) UnmarshalJSON(v []byte) error {
	*b = v
	return nil
}

func (b *Bytes) MarshalBSON() ([]byte, error) {
	if b == nil || len(*b) == 0 {
		return []byte("\"\""), nil
	}
	return *b, nil
}
func (b *Bytes) UnmarshalBSON(v []byte) error {
	*b = v
	return nil
}

func (b *Bytes) Marshal(v interface{}) error {
	if v == nil {
		return nil
	}
	d, err := json.Marshal(v)
	if err == nil {
		*b = Bytes(d)
	}
	return err
}

func (b *Bytes) Unmarshal(i interface{}) error {
	if len(*b) == 0 {
		return nil
	}
	return json.Unmarshal(*b, i)
}
