package binder

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"io"
)

func init() {
	b := &protobufBinding{}
	_ = Register(EncodingTypeProtoBuf, b)
}

type protobufBinding struct{}

func (*protobufBinding) String() string {
	return "protobuf"
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
