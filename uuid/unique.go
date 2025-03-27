package uuid

import (
	"fmt"
	"github.com/hwcer/cosgo/random"
	"strings"
	"sync/atomic"
	"time"
)

// Unique 不需要设置自增种子
// 比Builder使用简单
// 但是只能生成比较长的字符串
type Unique struct {
	base    int
	shard   string
	index   uint64
	suffix  string
	Garbled int //乱码长度，默认无乱码
}

func NewUnique(shard uint64, base int) *Unique {
	u := &Unique{base: base}
	u.shard = Pack(shard, base)
	t, _ := time.Parse("2006-01-02 15:04:05-0700", "2024-04-11 12:00:00+0800")
	v := time.Now().Unix() - t.Unix()
	u.suffix = Pack(uint64(v), base)
	return u
}

func (u *Unique) New(prefix uint64) string {
	i := atomic.AddUint64(&u.index, 1)
	var build strings.Builder
	build.WriteString(u.shard)
	build.WriteString(Pack(prefix, u.base))
	build.WriteString(u.suffix)
	build.WriteString(Pack(i, u.base))
	if u.Garbled > 0 {
		build.WriteString(fmt.Sprintf("%d", u.Garbled))
		build.WriteString(random.Strings.String(u.Garbled))
	}
	return build.String()
}

func (u *Unique) Simple() string {
	i := atomic.AddUint64(&u.index, 1)
	var build strings.Builder
	build.WriteString(u.suffix)
	build.WriteString(Pack(i, u.base))
	return build.String()
}
