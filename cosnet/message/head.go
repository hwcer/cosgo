package message

import (
	"bytes"
	"errors"
	"github.com/hwcer/cosgo/utils"
)

//TCP MESSAGE

const HeadSize = 9

var DefaultContentType = ContentTypeJson

type ContentType uint8

const (
	ContentTypeNumber ContentType = 1
	ContentTypeString ContentType = 2
	ContentTypeJson   ContentType = 3
	ContentTypeXml    ContentType = 4
	ContentTypeProto  ContentType = 5
)

type Head struct {
	Size        int32       //数据BODY长度 4294967295 4
	Code        uint16      //协议号 2
	Flags       Flags       //1
	contentType ContentType //1
	attachIndex AttachIndex //1
}

//parse 解析[]byte并填充字段
func (h *Head) Parse(head []byte) error {
	if len(head) != HeadSize {
		return errors.New("head len error")
	}
	utils.BytesToInt(head[0:4], &h.Size)
	utils.BytesToInt(head[4:6], &h.Code)
	utils.BytesToInt(head[6:7], &h.Flags)
	utils.BytesToInt(head[7:8], &h.contentType)
	utils.BytesToInt(head[8:9], &h.attachIndex)
	return nil
}

//Bytes 生成成byte类型head
func (h *Head) Bytes() []byte {
	var b [][]byte
	b = append(b, utils.IntToBytes(h.Size))
	b = append(b, utils.IntToBytes(h.Code))
	b = append(b, utils.IntToBytes(h.Flags))
	b = append(b, utils.IntToBytes(h.contentType))
	b = append(b, utils.IntToBytes(h.attachIndex))
	return bytes.Join(b, []byte{})
}
