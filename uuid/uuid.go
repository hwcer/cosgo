package uuid

import (
	"strconv"
	"strings"
)

const BaseSize = 36

type UUID struct {
	share  uint64
	prefix uint64
	suffix uint64
}

func (u *UUID) GetShard() uint64 {
	return u.share
}
func (u *UUID) GetPrefix() uint64 {
	return u.prefix
}
func (u *UUID) GetSuffix() uint64 {
	return u.suffix
}

// New 通过改变 prefix 生成新UUID
func (u *UUID) New(prefix uint64) *UUID {
	n := *u
	n.prefix = prefix
	return &n
}

func (u *UUID) String(base int) string {
	var build strings.Builder
	build.WriteString(Pack(u.share, base))
	build.WriteString(Pack(u.prefix, base))
	build.WriteString(Pack(u.suffix, base))
	return build.String()
}

// Uint64 转换成UINT64 可能超过math.MaxUint64
func (u *UUID) Uint64() (r uint64, err error) {
	s := u.String(10)
	r, err = strconv.ParseUint(s, 10, 64)
	return
}

func (u *UUID) Parse(id string, base int) (err error) {
	var i uint64
	suffix := id

	if i, suffix, err = Split(suffix, base, 0); err != nil {
		return
	} else {
		u.share = i
	}

	if i, suffix, err = Split(suffix, base, 0); err != nil {
		return
	} else {
		u.prefix = i
	}
	if i, suffix, err = Split(suffix, base, 0); err != nil {
		return
	} else {
		u.suffix = i
	}
	return nil
}
