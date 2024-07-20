package apis

import "strings"

// 接口权限设置

const (
	None      int8 = iota //不需要登录
	OAuth                 //需要认证
	Character             //需要选择角色
)

var Default = Character

var apis = map[string]int8{}

func Set(s string, i int8) {
	s = strings.ToLower(s)
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	apis[s] = i
}

func Get(s string) int8 {
	s = strings.ToLower(s)
	if !strings.HasPrefix(s, "/") {
		s = "/" + s
	}
	if v, ok := apis[s]; !ok {
		return Default
	} else {
		return v
	}
}

func Range(f func(s string, i int8)) {
	s := make(map[string]int8, len(apis))
	for k, v := range apis {
		s[k] = v
	}
	for k, v := range s {
		f(k, v)
	}
}
