package binder

import (
	"gopkg.in/yaml.v2"
	"io"
)

func init() {
	b := yamlBinding{}
	_ = Register(EncodingTypeYaml, b)
}

type yamlBinding struct{}

func (yamlBinding) String() string {
	return "yaml"
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
