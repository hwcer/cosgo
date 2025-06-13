package binder

import (
	"encoding/json"
	"io"
)

var Json = jsonBinding{}

func init() {
	_ = Register(MIMEJSON, Json)
}

type jsonBinding struct{}

func (jsonBinding) Id() uint8 {
	return Type(MIMEJSON).Id
}

func (jsonBinding) Name() string {
	return Type(MIMEJSON).Name
}
func (jsonBinding) String() string {
	return MIMEJSON
}
func (jsonBinding) Encode(w io.Writer, i interface{}) error {
	return json.NewEncoder(w).Encode(i)
}

func (jsonBinding) Decode(r io.Reader, i interface{}) error {
	return json.NewDecoder(r).Decode(i)
}

func (jsonBinding) Marshal(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

func (jsonBinding) Unmarshal(b []byte, i interface{}) error {
	if len(b) == 0 {
		return nil
	}
	return json.Unmarshal(b, i)
}
