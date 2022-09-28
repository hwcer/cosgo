// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"encoding/json"
	"fmt"
	"io"
)

func init() {
	j := jsonBinding{}
	_ = Register(MIMEJSON, j)
}

type jsonBinding struct{}

func (jsonBinding) Name() string {
	return "json"
}

func (jsonBinding) Bind(body io.Reader, obj interface{}) error {
	if body == nil {
		return fmt.Errorf("invalid request")
	}
	return json.NewDecoder(body).Decode(obj)
}

func (jsonBinding) Unmarshal(b []byte, obj interface{}) error {
	return json.Unmarshal(b, obj)
}
