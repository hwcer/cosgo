package uuid

import (
	"strconv"
	"strings"
)

type UUID struct {
	share  uint16
	prefix uint32
	suffix uint32
}

func (u *UUID) GetShard() uint16 {
	return u.share
}
func (u *UUID) GetPrefix() uint32 {
	return u.prefix
}
func (u *UUID) GetSuffix() uint32 {
	return u.suffix
}

// New 通过改变 prefix 生成新UUID
func (u *UUID) New(prefix uint32) *UUID {
	n := *u
	n.prefix = prefix
	return &n
}

func (u *UUID) String(base int) string {
	var build strings.Builder
	build.WriteString(Pack(uint32(u.share), base))
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
	var i uint32
	suffix := id

	if i, suffix, err = Split(suffix, base, 0); err != nil {
		return
	} else {
		u.share = uint16(i)
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
