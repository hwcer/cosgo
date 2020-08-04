package confer

import (
	"bytes"
	"common/handler"
	"fmt"
	"github.com/spf13/viper"
	"icefire/values"
	"log"
	"sync"
)

type confdConfer struct {
	baseConfer
	Endpoints []string
	Version   int64
	Value     string
	Kmps      map[string]ValChgHandler
	Mutx      sync.RWMutex
}

type confdGetR struct {
	Code    int    `json:"code"`
	Msg     string `json:"msg"`
	Data    []byte `json:"data"`
	Version int64  `json:"version"`
}

func newConfdConfer(provider string, param ConferParam) *confdConfer {
	v := new(confdConfer)
	v.Viper = *viper.New()
	v.Body = []byte{}

	v.ProviderType = provider
	v.Name = param.Name
	v.BodyType = param.Typ
	v.Endpoints = param.Addr
	v.Kmps = make(map[string]ValChgHandler)
	return v
}

func (this *confdConfer) Verify() error {
	return this.VerifyExt()
}

func (this *confdConfer) Read(val interface{}) error {
	// TODO try
	var err error

	url := fmt.Sprintf("%v/etc/get", values.GetAppPlatAddr())
	query := make(map[string]string)
	query["ServerId"] = this.Name
	reply := handler.OAuth.PostJson(url, query)
	if reply.HasError() {
		return fmt.Errorf("confd resp inv code:%v,msg:%v", reply.Code, reply.Error)
	}

	this.Body = reply.Data.([]byte)
	//logger.DEBUG("remote config:%v\n", string(this.Body))
	//this.Version = jresp.Version

	//log.Println("confd read conf:", string(this.Body))
	data := this.Body
	if this.OrigType == DEF_HCONF_TYPE_HJSON || this.OrigType == DEF_HCONF_TYPE_LUA {
		err, data = this.ToJson(this.OrigType, data)
		if err != nil {
			return err
		}
		this.BodyType = DEF_HCONF_TYPE_JSON
	}
	this.SetConfigType(this.BodyType)

	err = this.ReadConfig(bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("confer remote reader failed:%v", err)
	}

	if val != nil {
		err := this.Unmarshal(val)
		if err != nil {
			return fmt.Errorf("confer remote unmarshal failed:%v", err)
		}
	}

	log.Printf(">> READ CONFD")

	//go this.Watch()
	return nil
}

func (this *confdConfer) Watch() {
	// TODO
}

func (this *confdConfer) Register(key string, handler ValChgHandler) error {
	this.Mutx.Lock()
	defer this.Mutx.Unlock()
	if this.IsSet(key) == false {
		return fmt.Errorf("key not found")
	}

	this.Kmps[key] = handler
	return nil
}
