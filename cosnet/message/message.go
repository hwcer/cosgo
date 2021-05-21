package message

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/hwcer/cosgo/utils"
)

type Message struct {
	Head   *Head   //消息头，可能为nil
	Data   []byte  //消息数据
	Attach *Attach //自定义附件
}

func New(code uint16, body interface{}, dataType ContentType) *Message {
	msg := &Message{Head: &Head{Code: code, DataType: dataType}, Attach: NewAttach()}
	switch dataType {
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
	msg.Head.Size = int32(len(msg.Data))
	return msg
}

//通过 Head bytes 创建msg
func NewMsg(head []byte) (*Message, error) {
	msg := &Message{Head: &Head{}, Attach: NewAttach()}
	if err := msg.Head.Parse(head); err != nil {
		return nil, err
	}
	return msg, nil
}

//Bytes 生成二进制文件
func (r *Message) Bytes() []byte {
	var b [][]byte
	b = append(b, r.Head.Bytes())
	if len(r.Data) > 0 {
		b = append(b, r.Data)
	}
	return bytes.Join(b, []byte{})
}

func (r *Message) Bind(i interface{}) error {
	dt := r.Head.DataType
	if dt == 0 {
		dt = MsgDataTypeDefault
	}
	switch dt {
	case ContentTypeJson:
		return json.Unmarshal(r.Data, i)
	case ContentTypeProto:
		return proto.Unmarshal(r.Data, i.(proto.Message))
	case ContentTypeXml:
		return xml.Unmarshal(r.Data, i)
	default:
		return errors.New("unknown ContentType")
	}
}
