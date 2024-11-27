package uuid

import (
	"strings"
	"sync/atomic"
	"time"
)

// Unique 不需要设置自增种子
// 比Builder使用简单
// 但是只能生成比较长的字符串
type Unique struct {
	base   int
	shard  string
	index  uint32
	suffix string
}

func NewUnique(shard uint32, base int) *Unique {
	u := &Unique{base: base}
	u.shard = Pack(shard, base)
	t, _ := time.Parse("2006-01-02 15:04:05-0700", "2024-04-11 12:00:00+0800")
	v := time.Now().Unix() - t.Unix()
	u.suffix = Pack(uint32(v), base)
	return u
}

func (u *Unique) New(prefix uint32) string {
	i := atomic.AddUint32(&u.index, 1)
	var build strings.Builder
	build.WriteString(u.shard)
	build.WriteString(Pack(prefix, u.base))
	build.WriteString(u.suffix)
	build.WriteString(Pack(i, u.base))
	return build.String()
}

func (u *Unique) Simple() string {
	i := atomic.AddUint32(&u.index, 1)
	var build strings.Builder
	build.WriteString(u.suffix)
	build.WriteString(Pack(i, u.base))
	return build.String()
}
