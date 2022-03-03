package request

import (
	"encoding/json"
)

type Packer interface {
	Encode(i interface{}) ([]byte, error)
	Decode(b []byte, i interface{}) error
	ContentType() string
}

type PackerJson struct{}

func (b *PackerJson) Encode( i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

func (b *PackerJson) Decode( data []byte, i interface{}) (err error) {
	return json.Unmarshal(data,i)
}

func (b *PackerJson) ContentType() string {
	return "application/json;charset=utf-8"
}
