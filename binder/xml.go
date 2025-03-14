package binder

import (
	"encoding/xml"
	"io"
)

var Xml = &xmlBinding{}

func init() {
	_ = Register(MIMEXML, Xml)
	_ = Register(MIMEXML2, Xml)
}

type xmlBinding struct{}

func (xmlBinding) Id() uint8 {
	return Type(MIMEXML).Id
}

func (xmlBinding) Name() string {
	return Type(MIMEXML).Name
}
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
