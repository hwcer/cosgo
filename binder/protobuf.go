package binder

import (
	"errors"
	"google.golang.org/protobuf/proto"
	"io"
)

var Protobuf = &protobufBinding{}

func init() {
	_ = Register(MIMEPROTOBUF, Protobuf)
}

type protobufBinding struct{}

func (*protobufBinding) Id() uint8 {
	return Type(MIMEPROTOBUF).Id
}

func (*protobufBinding) Name() string {
	return Type(MIMEPROTOBUF).Name
}
func (*protobufBinding) String() string {
	return MIMEPROTOBUF
}
func (this *protobufBinding) Encode(w io.Writer, i interface{}) error {
	b, err := this.Marshal(i)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
}

func (this *protobufBinding) Decode(body io.Reader, obj interface{}) error {
	buf, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	return this.Unmarshal(buf, obj)
}

func (*protobufBinding) Marshal(i interface{}) ([]byte, error) {
	if i == nil {
		return []byte{}, nil
	}
	pb, ok := i.(proto.Message)
	if !ok {
		return nil, errors.New("not proto.Message")
	}
	return proto.Marshal(pb)
}

func (*protobufBinding) Unmarshal(b []byte, i interface{}) error {
	pb, ok := i.(proto.Message)
	if !ok {
		return errors.New("not proto.Message")
	}
	return proto.Unmarshal(b, pb)
}
