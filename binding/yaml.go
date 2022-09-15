// Copyright 2018 Gin Core Team.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"gopkg.in/yaml.v2"
	"io"
)

func init() {
	b := yamlBinding{}
	Register(MIMEYAML, b)
}

type yamlBinding struct{}

func (yamlBinding) Name() string {
	return "yaml"
}

func (yamlBinding) Bind(body io.Reader, obj interface{}) error {
	decoder := yaml.NewDecoder(body)

	return decoder.Decode(obj)
}
func (yamlBinding) Unmarshal(b []byte, obj interface{}) error {
	return yaml.Unmarshal(b, obj)
}
