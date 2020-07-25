package handler

import (
	"github.com/golang/protobuf/proto"
)

type MessageDataType int

const (
	MessageDataTypeString MessageDataType = iota
	MessageDataTypeJson
	MessageDataTypeXml
	MessageDataTypeProtobuf
)



//http 回包信息
type Message struct {
	Code     int       		//错误码
	Data     interface{}    //回包信息
	Error    string   		//错误描述，一般是系统错误
	DataType MessageDataType   		//Data TYPE
	Status int                      //http Status
}

func (m *Message)HasError()bool  {
	return m.Code != GetErrCode(ErrMsg_NAME_SUCCESS)
}


func NewMsgReply(data interface{},args ...MessageDataType) *Message {
	msg :=  &Message{Code: GetErrCode(ErrMsg_NAME_SUCCESS)}
	msg.Data = data
	if len(args) > 0{
		msg.DataType = args[0]
	}
	return msg
}

func NewErrMsgReply(err string,code ...int) *Message {
	e := &Message{Error: err}
	if len(code)>0{
		e.Code = code[0]
	}else {
		e.Code = GetErrCode(err)
	}
	return e
}

func NewMsgError(err error,code ...int) *Message {
	return NewErrMsgReply(err.Error(),code...)
}
//JSON
func NewXmlMsgReply(v interface{}) *Message {
	return NewMsgReply(v,MessageDataTypeXml)
}
//JSON
func NewJsonMsgReply(v interface{}) *Message {
	return NewMsgReply(v,MessageDataTypeJson)
}
//string
func NewStringMsgReply(v string) *Message {
	return NewMsgReply(v,MessageDataTypeString)
}
//protobuf
func NewProtoMsgReply(v proto.Message) *Message {
	return NewMsgReply(v,MessageDataTypeProtobuf)
}
