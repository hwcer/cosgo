package confer

import (
	"time"
)

const (
	DEF_HCONF_PROVIDER_LOCAL = "PLOCAL" // 本地配置
	DEF_HCONF_PROVIDER_CONFD = "PCONFD" // 自定义配置中心

	DEF_HCONF_TYPE_JSON  = "json"
	DEF_HCONF_TYPE_HJSON = "hjson"
	DEF_HCONF_TYPE_LUA   = "lua"
)

type ConferParam struct {
	Name string // 本地文件名 或者 远程配置的路径，如不包括扩展名，需要通过Typ 指定
	Typ  string // 配置格式类型

	Path   string   // 本地文件路径
	Addr   []string // 远程配置读取地址
	Buffer []byte   //
}

type ValChgHandler func(ov, nv interface{})

type Confer interface {
	IsSet(key string) bool
	InConfig(key string) bool
	AllKeys() []string

	Get(key string) interface{}
	GetBool(key string) bool
	GetFloat64(key string) float64
	GetInt(key string) int

	GetString(key string) string
	GetStringMap(key string) map[string]interface{}
	GetStringMapString(key string) map[string]string
	GetStringSlice(key string) []string
	GetStringMapStringSlice(key string) map[string][]string
	GetSizeInBytes(key string) uint
	GetTime(key string) time.Time
	GetDuration(key string) time.Duration

	Set(key string, value interface{})

	UnmarshalByKey(key string, rawVal interface{}) error

	Verify() error
	Read(val interface{}) error
	Register(key string, handler ValChgHandler) error
}
