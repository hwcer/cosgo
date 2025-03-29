package uuid

import (
	"fmt"
	"math"
	"sync/atomic"
)

type Builder struct {
	shard uint64
	index uint64 //UID递增ID
}

// New 创建种子,使用自增种子，但初始化时需要手动设置当前种子
//
//	shard 服务器分片ID
//	index 自增种子,如果不使用UUID可以为0
func New(shard uint64, index uint64) *Builder {
	if index >= math.MaxUint32 {
		panic("uuid index out of range")
	}
	u := &Builder{}
	u.shard = shard
	u.index = index
	return u
}

func Create(id any, base int) (*Builder, error) {
	s := fmt.Sprintf("%v", id)
	u := &UUID{}
	if err := u.Parse(s, base); err != nil {
		return nil, err
	}
	return New(u.share, u.index), nil
}

func (u *Builder) Shard() uint64 {
	return u.shard
}

func (u *Builder) Index() uint64 {
	return u.index
}

// New 生成UUID
func (u *Builder) New(prefix uint64) *UUID {
	r := &UUID{}
	r.share = u.shard
	r.prefix = prefix
	r.index = atomic.AddUint64(&u.index, 1)
	return r
}
