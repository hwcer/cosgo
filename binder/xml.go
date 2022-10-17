package binder

import (
	"encoding/xml"
	"io"
)

func init() {
	b := xmlBinding{}
	_ = Register(MIMEXML, b)
	_ = Register(MIMEXML2, b)
}

type xmlBinding struct{}

func (xmlBinding) String() string {
	return MIMEXML
}
func (xmlBinding) Encode(w io.Writer, i interface{}) error {
	return xml.NewEncoder(w).Encode(i)
}

func (xmlBinding) Decode(r io.Reader, i interface{}) error {
	return xml.NewDecoder(r).Decode(i)
}

func (xmlBinding) Marshal(i interface{}) ([]byte, error) {
	return xml.Marshal(i)
}

func (xmlBinding) Unmarshal(b []byte, i interface{}) error {
	return xml.Unmarshal(b, i)
}
