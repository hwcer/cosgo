// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"github.com/golang/protobuf/proto"
	"io"
)

func init() {
	b := protobufBinding{}
	Register(MIMEPROTOBUF, b)
}

type protobufBinding struct{}

func (protobufBinding) Name() string {
	return "protobuf"
}

func (b protobufBinding) Bind(body io.Reader, obj interface{}) error {
	buf, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	return b.Unmarshal(buf, obj)
}

func (protobufBinding) Unmarshal(body []byte, obj interface{}) error {
	if err := proto.Unmarshal(body, obj.(proto.Message)); err != nil {
		return err
	}
	// Here it's same to return validate(obj), but util now we can't add
	// `binding:""` to the struct which automatically generate by gen-proto
	return nil
	// return validate(obj)
}
