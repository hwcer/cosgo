// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"encoding/xml"
	"io"
)

func init() {
	b := xmlBinding{}
	Register(MIMEXML, b)
	Register(MIMEXML2, b)
}

type xmlBinding struct{}

func (xmlBinding) Name() string {
	return "xml"
}

func (xmlBinding) Bind(body io.Reader, obj interface{}) error {
	decoder := xml.NewDecoder(body)
	return decoder.Decode(obj)
}
func (xmlBinding) Unmarshal(b []byte, obj interface{}) error {
	return xml.Unmarshal(b, obj)
}
