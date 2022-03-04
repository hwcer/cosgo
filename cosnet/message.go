package cosnet

import (
	"bytes"
	"errors"
	"github.com/hwcer/cosgo/library/logger"
	"github.com/hwcer/cosgo/utils"
)

const HeaderSize = 8

var MaxMsgDataSize uint32 = 100 * 1024 //1M

type Message interface {
	Code() interface{}                      //协议号
	Size() uint32                           //BODY长度
	Data() []byte                           //获取包体
	Index() uint16                          //唯一编号
	Reset(data []byte, args ...interface{}) //设置包体
	Parse(head []byte) error                //解析二进制头并填充到对应字段
	Bytes() ([]byte, error)                 //转换成[]byte
}

type message struct {
	code  uint16 //协议号 2
	size  uint32 //数据BODY 4
	index uint16 //唯一编号 2
	data  []byte //消息数据
}

func (this *message) Code() interface{} {
	return this.code
}
func (this *message) Size() uint32 {
	return this.size
}

func (this *message) Data() []byte {
	return this.data
}

func (this *message) Index() uint16 {
	return this.index
}

//data code index
func (this *message) Reset(data []byte, args ...interface{}) {
	this.size = uint32(len(data))
	this.data = data
	if len(args) > 0 && args[0] != nil {
		this.code = args[0].(uint16)
	}
	if len(args) > 1 && args[1] != nil {
		this.index = args[1].(uint16)
	}
}

//Parse 解析二进制头并填充到对应字段
func (this *message) Parse(head []byte) error {
	if len(head) != HeaderSize {
		return errors.New("head len error")
	}
	if err := utils.BytesToInt(head[0:2], &this.code); err != nil {
		return err
	}
	if err := utils.BytesToInt(head[2:6], &this.size); err != nil {
		return err
	}
	if err := utils.BytesToInt(head[6:8], &this.index); err != nil {
		return err
	}
	if this.size > MaxMsgDataSize {
		logger.Debug("包体太长，可能是包头错误:%+v", this)
		return ErrMsgDataSizeTooLong
	}
	return nil
}

//Bytes 生成二进制文件
func (this *message) Bytes() ([]byte, error) {
	buffer := new(bytes.Buffer)
	if err := utils.IntToBuffer(buffer, this.code); err != nil {
		return nil, err
	}
	if err := utils.IntToBuffer(buffer, this.size); err != nil {
		return nil, err
	}
	if err := utils.IntToBuffer(buffer, this.index); err != nil {
		return nil, err
	}
	if this.size > 0 {
		buffer.Write(this.data)
	}
	return buffer.Bytes(), nil
}
