package storage

import (
	"errors"
	"github.com/hwcer/cosgo/await"
	"github.com/hwcer/cosgo/uuid"
)

type Cache struct {
	cap        int
	bucket     []*Bucket
	initialize *await.Initialize
	NewSetter  NewSetter //创建新数据结构
}

func (this *Cache) Share(id string) (int, error) {
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
func (this *Cache) Get(id string) (Setter, bool) {
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

func (this *Cache) Set(id string, v any) bool {
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
func (this *Cache) Size() (r int) {
	for _, bucket := range this.bucket {
		r += bucket.Size()
	}
	return
}

// Range 遍历
func (this *Cache) Range(f func(Setter) bool) bool {
	for _, bucket := range this.bucket {
		if !bucket.Range(f) {
			return false
		}
	}
	return true
}

// Create 创建一个新数据
func (this *Cache) Create(v any) Setter {
	for _, bucket := range this.bucket {
		if r := bucket.Create(v); r != nil {
			return r
		}
	}
	return this.expansion(v)
}
func (this *Cache) expansion(v any) Setter {
	i := len(this.bucket)
	_ = this.initialize.Reload(this.append)
	bucket := this.bucket[i]
	if r := bucket.Create(v); r != nil {
		return r
	}
	return this.expansion(v)
}

func (this *Cache) append() error {
	bucket := NewBucket(len(this.bucket), this.cap)
	bucket.NewSetter = this.NewSetter
	this.bucket = append(this.bucket, bucket)
	this.initialize = await.NewInitialize()
	return nil
}

// Delete 删除并返回已删除的数据
func (this *Cache) Delete(id string) Setter {
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
