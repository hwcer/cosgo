package confer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/mitchellh/mapstructure"
	"path"
	"strings"

	"github.com/hjson/hjson-go"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
)

// 支持的配置格式
var defConfExts = []string{"json", "lua", "hjson", "toml", "yaml", "yml", "properties", "props", "prop", "hcl"}

type baseConfer struct {
	viper.Viper

	Name         string
	Body         []byte
	BodyType     string
	OrigType     string
	ProviderType string
	Val          interface{}
}

func (this *baseConfer) VerifyExt() error {
	if this.BodyType == "" {
		ext := path.Ext(this.Name)
		if ext == "" {
			return fmt.Errorf("confer verify get null ext")
		}
		this.BodyType = ext[1:]
	}

	this.BodyType = strings.ToLower(this.BodyType)
	ok := false
	for _, t := range defConfExts {
		if this.BodyType == t {
			ok = true
			break
		}
	}

	if ok == false {
		return fmt.Errorf("confer unsupport body type:%v", this.BodyType)
	}

	this.OrigType = this.BodyType
	return nil
}

func (this *baseConfer) ToJson(typ string, data []byte) (error, []byte) {
	switch typ {
	case DEF_HCONF_TYPE_HJSON:
		// hjson to json
		var hval interface{}
		if err := hjson.Unmarshal(data, &hval); err != nil {
			return err, nil
		}
		dt, err := json.Marshal(hval)
		if err != nil {
			return err, nil
		}
		return nil, fixJSON(dt)

	case DEF_HCONF_TYPE_LUA:
		// lua to json
		lsrc := string(data)
		dt, err := luaToJson(lsrc)
		if err != nil {
			return err, nil
		}
		return nil, []byte(dt)

	default:
		return fmt.Errorf("unsupport fmt:%v", typ), nil
	}
}

func (this *baseConfer) Marshal(typ string, indent bool) ([]byte, error) {
	var rawVal interface{}

	err := this.Unmarshal(&rawVal)
	if err != nil {
		return nil, err
	}

	switch typ {
	case DEF_HCONF_TYPE_JSON:
		if indent == true {
			br, er := json.MarshalIndent(&rawVal, "", "  ")
			return br, er
		}
		br, er := json.Marshal(&rawVal)
		return br, er

	case DEF_HCONF_TYPE_HJSON:
		opt := hjson.DefaultOptions()
		opt.BracesSameLine = true
		br, er := hjson.MarshalWithOptions(rawVal, opt)
		return br, er

	default:
		return nil, fmt.Errorf("invalid marshal type:%v", typ)
	}
}

func decodeOption(dc *mapstructure.DecoderConfig) {
	dc.TagName = "json"
}

func (this *baseConfer) UnmarshalByKey(key string, rawVal interface{}) error {
	return this.Viper.UnmarshalKey(key, rawVal, decodeOption)
}

func fixJSON(data []byte) []byte {
	data = bytes.Replace(data, []byte("\\u003c"), []byte("<"), -1)
	data = bytes.Replace(data, []byte("\\u003e"), []byte(">"), -1)
	data = bytes.Replace(data, []byte("\\u0026"), []byte("&"), -1)
	data = bytes.Replace(data, []byte("\\u0008"), []byte("\\b"), -1)
	data = bytes.Replace(data, []byte("\\u000c"), []byte("\\f"), -1)
	return data
}
