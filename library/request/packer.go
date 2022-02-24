package request

import (
	"encoding/json"
	"io"
)

type Packer interface {
	Encode(b io.Writer, i interface{}) error
	Decode(b io.Reader, i interface{}) error
	ContentType() string
}

type PackerJson struct{}

func (b *PackerJson) Encode(w io.Writer, i interface{}) (err error) {
	return json.NewEncoder(w).Encode(i)
}

func (b *PackerJson) Decode(r io.Reader, i interface{}) (err error) {
	return json.NewDecoder(r).Decode(i)
}

func (b *PackerJson) ContentType() string {
	return "application/json;charset=utf-8"
}
