package confer

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

type localConfer struct {
	baseConfer

	LocalDir string
}

func newLocalConfer(provider string, param ConferParam) *localConfer {
	v := new(localConfer)
	v.Viper = *viper.New()
	v.Body = []byte{}

	v.ProviderType = provider
	v.Name = param.Name
	v.BodyType = param.Typ

	v.LocalDir = param.Path
	return v
}

func (this *localConfer) Verify() error {
	err := this.VerifyExt()
	if err != nil {
		return err
	}

	if filepath.IsAbs(this.Name) == false && this.LocalDir == "" {
		return fmt.Errorf("confer local invalid path")
	}

	return nil
}

func (this *localConfer) Read(val interface{}) error {
	// read
	filePath := this.Name
	if filepath.IsAbs(filePath) == false {
		sd, sf := filepath.Split(filePath)
		filePath = filepath.Join(this.LocalDir, sd, sf)
	}
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// json bom fix
	if this.OrigType == DEF_HCONF_TYPE_JSON {
		data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
	}

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
		return fmt.Errorf("confer local reader failed:%v", err)
	}

	if val != nil {
		err := this.Unmarshal(val)
		if err != nil {
			return fmt.Errorf("confer local unmarshal failed:%v", err)
		}
	}

	return nil
}

func (this *localConfer) Register(key string, handler ValChgHandler) error {
	return fmt.Errorf("confer local not support")
}
