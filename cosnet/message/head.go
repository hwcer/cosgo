package message

import (
	"bytes"
	"errors"
	"github.com/hwcer/cosgo/utils"
)

//TCP MESSAGE

var MsgHeadSize = 8
var MsgDataTypeDefault ContentType = ContentTypeJson

func init() {
	setAttachField(14, "Flags", 8)
	setAttachField(15, "ContentType", 8)
}

type Head struct {
	Size   int32   //数据BODY长度 4294967295 4
	Code   uint16  //协议号 2
	Attach *Attach //附件内容
}

//parse 解析[]byte并填充字段
func (m *Head) Parse(head []byte) error {
	if len(head) != MsgHeadSize {
		return errors.New("head len error")
	}
	utils.BytesToInt(head[0:4], &m.Size)
	utils.BytesToInt(head[4:6], &m.Code)
	utils.BytesToInt(head[6:8], &m.Length)
	return nil
}

//Bytes 生成成byte类型head
func (m *Head) Bytes() []byte {
	var b [][]byte
	b = append(b, utils.IntToBytes(m.Size))
	b = append(b, utils.IntToBytes(m.Code))

	attach := m.Attach.Bytes()
	if len(attach) > 0 {

	}
	b = append(b, utils.IntToBytes(m.Flags))
	b = append(b, utils.IntToBytes(m.DataType))
	return bytes.Join(b, []byte{})
}
