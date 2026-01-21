package storage

import (
	"errors"
	"sync"

	"github.com/hwcer/cosgo/uuid"
)

// New 创建一个新的存储实例
// cap: 每个桶的容量
// creator: 可选的自定义setter创建函数
func New(cap int, creator ...NewSetter) *Storage {
	r := &Storage{cap: cap}
	if len(creator) > 0 {
		r.NewSetter = creator[0]
	} else {
		r.NewSetter = NewSetterDefault
	}
	//创建第一个bucket
	bucket := NewBucket(len(r.bucket), r.cap)
	bucket.NewSetter = r.NewSetter
	r.bucket = append(r.bucket, bucket)
	return r
}

type Storage struct {
	cap       int
	bucket    []*Bucket
	NewSetter NewSetter //创建新数据结构
	mu        sync.RWMutex
}

func (this *Storage) Share(id string) (int, error) {
	i, _, err := uuid.Split(id, uuid.BaseSize, 0)
	if err != nil {
		return 0, nil
	}
	r := int(i)
	if r < 0 || r >= len(this.bucket) {
		return 0, errors.New("invalid id")
	}
	return r, nil
}
func (this *Storage) Get(id string) (Setter, bool) {
	share, err := this.Share(id)
	if err != nil {
		return nil, false
	}
	bucket := this.bucket[share]
	if bucket == nil {
		return nil, false
	}
	return bucket.Get(id)
}

func (this *Storage) Set(id string, v any) bool {
	share, err := this.Share(id)
	if err != nil {
		return false
	}
	bucket := this.bucket[share]
	if bucket == nil {
		return false
	}
	return bucket.Set(id, v)
}

// Size 当前数量
func (this *Storage) Size() (r int) {
	for _, bucket := range this.bucket {
		r += bucket.Size()
	}
	return
}

// Free 当前空闲
func (this *Storage) Free() (r int) {
	for _, bucket := range this.bucket {
		r += bucket.Free()
	}
	return
}

// Range 遍历
func (this *Storage) Range(f func(Setter) bool) bool {
	for _, bucket := range this.bucket {
		if !bucket.Range(f) {
			return false
		}
	}
	return true
}

// Create 创建一个新数据
func (this *Storage) Create(v any) Setter {
	for _, bucket := range this.bucket {
		if r := bucket.Create(v); r != nil {
			return r
		}
	}
	return this.expansion(v)
}
func (this *Storage) expansion(v any) Setter {
	this.mu.Lock()
	defer this.mu.Unlock()
	// 再次检查是否存在
	for _, bucket := range this.bucket {
		if r := bucket.Create(v); r != nil {
			return r
		}
	}
	// 扩展bucket
	bucket := NewBucket(len(this.bucket), this.cap)
	bucket.NewSetter = this.NewSetter
	r := bucket.Create(v)
	this.bucket = append(this.bucket, bucket)
	return r
}

// Delete 删除并返回已删除的数据
func (this *Storage) Delete(id string) Setter {
	share, err := this.Share(id)
	if err != nil {
		return nil
	}
	bucket := this.bucket[share]
	if bucket == nil {
		return nil
	}
	return bucket.Delete(id)
}
func (this *Storage) Remove(id []string) (r []Setter) {
	for _, v := range id {
		if s := this.Delete(v); s != nil {
			r = append(r, s)
		}
	}
	return r
}
