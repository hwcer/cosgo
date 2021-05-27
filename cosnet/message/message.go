package message

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/hwcer/cosgo/utils"
)

//TCP MESSAGE

type Message struct {
	Head   *Head   //消息头，可能为nil
	Data   []byte  //消息数据
	Attach *Attach //附件内容1
}

func New(code uint16, body interface{}, contentType ContentType) *Message {
	msg := &Message{Head: &Head{Code: code}, Attach: NewAttach(0)}
	switch contentType {
	case ContentTypeNumber:
		msg.Data = utils.IntToBytes(body)
	case ContentTypeString:
		msg.Data = []byte(body.(string))
	case ContentTypeJson:
		msg.Data, _ = json.Marshal(body)
	case ContentTypeXml:
		msg.Data, _ = xml.Marshal(body)
	case ContentTypeProto:
		msg.Data, _ = proto.Marshal(body.(proto.Message))
	}
	msg.Head.contentType = contentType
	msg.Head.Size = int32(len(msg.Data))
	return msg
}

//NewMsg 通过 Head bytes 创建msg
func NewMsg(head []byte) (*Message, error) {
	msg := &Message{Head: &Head{}}
	if err := msg.Head.Parse(head); err != nil {
		return nil, err
	}
	msg.Attach = NewAttach(msg.Head.attachIndex)
	return msg, nil
}

//Bytes 生成二进制文件
func (r *Message) Bytes() []byte {
	var b [][]byte
	r.Head.attachIndex = r.Attach.index
	b = append(b, r.Head.Bytes())
	if r.Attach.Size() > 0 {
		b = append(b, r.Attach.Bytes())
	}
	if len(r.Data) > 0 {
		b = append(b, r.Data)
	}
	return bytes.Join(b, []byte{})
}

func (r *Message) Bind(i interface{}) error {
	contentType := r.Head.contentType
	if contentType == 0 {
		contentType = DefaultContentType
	}
	switch contentType {
	case ContentTypeJson:
		return json.Unmarshal(r.Data, i)
	case ContentTypeProto:
		return proto.Unmarshal(r.Data, i.(proto.Message))
	case ContentTypeXml:
		return xml.Unmarshal(r.Data, i)
	default:
		return errors.New("unknown contentType")
	}
}

func (r *Message) ContentType() ContentType {
	return r.Head.contentType
}
