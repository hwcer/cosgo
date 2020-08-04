package server

import (
	"github.com/golang/protobuf/proto"
	"net/http"
)


type MessageDataType int

const (
	MessageDataType_String MessageDataType = iota
	MessageDataType_Json
	MessageDataType_Xml
	MessageDataType_Protobuf
	MessageDataType_Unknown  //输出完整的Message信息
)


const (
	MessageFlage_DataType 	 string = "_MessageFlage_DataType"
	MessageFlage_HttpStatus  string = "_MessageFlage_HttpStatus"
)




//http 回包信息
type Message struct {
	Code     int              //错误码
	Data     interface{}      //回包信息
	Error    string           //错误描述，一般是系统错误
	flags    map[string]interface{}  //标记参数
}

func (m *Message)HasError()bool  {
	return m.Code != GetErrCode(ErrMsg_NAME_SUCCESS)
}

func (m *Message)SetFlags(k string,v interface{})  {
	if m.flags == nil{
		m.flags = make(map[string]interface{})
	}
	m.flags[k] = v
}

func (m *Message)GetFlags(k string) (v interface{},ok bool)  {
	if m.flags == nil{
		return nil,false
	}
	v,ok= m.flags[k]
	return
}


func (m *Message)SetHttpStatus(code int) {
	m.SetFlags(MessageFlage_HttpStatus,code)
}

func (m *Message)GetHttpStatus() int {
	v,ok := m.GetFlags(MessageFlage_HttpStatus)
	if !ok{
		return http.StatusOK
	}else{
		return v.(int)
	}
}

func (m *Message)SetDataType(dt MessageDataType) {
	m.SetFlags(MessageFlage_DataType,dt)
}

func (m *Message)GetDataType() MessageDataType {
	v,ok := m.GetFlags(MessageFlage_DataType)
	if !ok{
		return MessageDataType_String
	}else{
		return v.(MessageDataType)
	}
}





func NewMsg(data interface{},args ...MessageDataType) *Message {
	msg :=  &Message{Code: GetErrCode(ErrMsg_NAME_SUCCESS)}
	msg.Data = data
	if len(args) > 0{
		msg.SetDataType(args[0])
	}
	return msg
}

func NewErrMsg(err string,code ...int) *Message {
	e := &Message{Error: err}
	if len(code)>0{
		e.Code = code[0]
	}else {
		e.Code = GetErrCode(err)
	}
	return e
}

func NewErrMsgFromError(err error,code ...int) *Message {
	return NewErrMsg(err.Error(),code...)
}
//JSON
func NewXmlMsg(v interface{}) *Message {
	return NewMsg(v,MessageDataType_Xml)
}
//JSON
func NewJsonMsg(v interface{}) *Message {
	return NewMsg(v,MessageDataType_Json)
}
//string
func NewStringMsg(v string) *Message {
	return NewMsg(v,MessageDataType_String)
}
//protobuf
func NewProtoMsg(v proto.Message) *Message {
	return NewMsg(v,MessageDataType_Protobuf)
}
