package options

import (
	"strings"
)

// 接口权限设置

const (
	ApisTypeNone      int8 = iota //不需要登录
	ApisTypeOAuth                 //需要认证
	ApisTypeCharacter             //需要选择角色
)

var Apis = apis{}

type apis map[string]int8

func init() {
	s := map[string]int8{
		"/login":       ApisTypeNone,
		"/role/create": ApisTypeOAuth,
		"/role/select": ApisTypeOAuth,
	}
	for k, v := range s {
		Apis.Set(k, v)
	}
}

func (auth apis) Set(s string, i int8) {
	s = strings.ToLower(s)
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	auth[s] = i
}

func (auth apis) Get(s string) int8 {
	s = strings.ToLower(s)
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	if v, ok := auth[s]; !ok {
		return ApisTypeCharacter
	} else {
		return v
	}
}

func (auth apis) Range(f func(s string, i int8)) {
	for k, v := range auth {
		f(k, v)
	}
}
