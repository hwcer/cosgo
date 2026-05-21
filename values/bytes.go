package values

import (
	"encoding/json"

	"go.mongodb.org/mongo-driver/v2/bson"
)

// Deprecated: Bytes 已废弃
type Bytes []byte

const (
	BytesEmptyJson = "null"
)

func (b *Bytes) MarshalJSON() ([]byte, error) {
	if b == nil || len(*b) == 0 {
		return []byte(BytesEmptyJson), nil
	}
	var d bson.D
	if err := bson.Unmarshal(*b, &d); err != nil {
		return *b, nil
	}
	return json.Marshal(d)
}

func (b *Bytes) UnmarshalJSON(v []byte) error {
	if string(v) == BytesEmptyJson {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(v, &m); err != nil {
		*b = append((*b)[:0], v...)
		return nil
	}
	data, err := bson.Marshal(m)
	if err != nil {
		return err
	}
	*b = data
	return nil
}

func (b *Bytes) MarshalBSONValue() (bson.Type, []byte, error) {
	if b == nil || len(*b) == 0 {
		return bson.TypeNull, nil, nil
	}
	return bson.MarshalValue(bson.Binary{Subtype: bson.TypeBinaryGeneric, Data: *b})
}

func (b *Bytes) UnmarshalBSONValue(typ bson.Type, data []byte) error {
	if typ == bson.TypeNull || len(data) == 0 {
		return nil
	}
	var bin bson.Binary
	if err := bson.UnmarshalValue(typ, data, &bin); err != nil {
		return err
	}
	*b = bin.Data
	return nil
}

func (b *Bytes) Marshal(v any) error {
	if v == nil {
		return nil
	}
	d, err := bson.Marshal(v)
	if err == nil {
		*b = d
	}
	return err
}

func (b *Bytes) Unmarshal(i any) error {
	if len(*b) == 0 {
		return nil
	}
	return bson.Unmarshal(*b, i)
}
