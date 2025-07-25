package uuid

import (
	"errors"
	"strconv"
	"strings"
)

func Pack(id uint64, base int) string {
	arr := make([]string, 2)
	if id != 0 {
		arr[1] = strconv.FormatUint(id, base)
	}
	arr[0] = strconv.FormatUint(uint64(len(arr[1])), base)
	return strings.Join(arr, "")
}

// Split 分割uuid
// index 取出第几段， 0开始
func Split(s string, base int, index int) (uint64, string, error) {
	var v string
	var p string
	p = s
	for i := 0; i <= index; i++ {
		x, err := Index(p, base)
		if err != nil {
			return 0, "", err
		}
		v = p[1:x]
		p = p[x:]
	}
	if v == "" {
		return 0, p, nil
	}
	if r, err := strconv.ParseUint(v, base, 63); err != nil {
		return 0, "", err
	} else {
		return r, p, nil
	}
}

// Index 获取有效字符串长度
func Index(id string, base int) (r int, err error) {
	var v int64
	if v, err = strconv.ParseInt(id[0:1], base, 64); err != nil {
		return
	} else {
		r = int(v) + 1
	}
	if r > len(id) {
		err = errors.New("oid error")
	}
	return
}

// IsValid 验证UUID 是否合法
func IsValid(uuid string, base ...int) bool {
	i := UUID{}
	b := int(0)
	if len(base) > 0 {
		b = base[0]
	} else {
		b = BaseSize
	}
	if err := i.Parse(uuid, b); err != nil {
		return false
	}
	return true
}
