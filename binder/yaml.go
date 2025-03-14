package binder

import (
	"gopkg.in/yaml.v2"
	"io"
)

func init() {
	b := yamlBinding{}
	_ = Register(MIMEYAML, b)
}

type yamlBinding struct{}

func (yamlBinding) Id() uint8 {
	return Type(MIMEYAML).Id
}

func (yamlBinding) Name() string {
	return Type(MIMEYAML).Name
}
func (yamlBinding) String() string {
	return MIMEYAML
}
func (yamlBinding) Encode(w io.Writer, i interface{}) error {
	return yaml.NewEncoder(w).Encode(i)
}

func (yamlBinding) Decode(r io.Reader, i interface{}) error {
	return yaml.NewDecoder(r).Decode(i)
}

func (yamlBinding) Marshal(i interface{}) ([]byte, error) {
	return yaml.Marshal(i)
}

func (yamlBinding) Unmarshal(b []byte, obj interface{}) error {
	return yaml.Unmarshal(b, obj)
}
