package storage

import (
	"errors"
	"sync"
	"sync/atomic"
)

// New 创建一个新的存储实例
// cap: 每个桶的容量
// creator: 可选的自定义 Setter 创建函数
func New(cap int, creator ...NewSetter) *Storage {
	r := &Storage{cap: cap}
	if len(creator) > 0 {
		r.NewSetter = creator[0]
	} else {
		r.NewSetter = NewSetterDefault
	}
	bucket := NewBucket(0, r.cap)
	bucket.NewSetter = r.NewSetter
	buckets := []*Bucket{bucket}
	r.buckets.Store(&buckets)
	r.totalCap.Store(int64(cap))
	return r
}

// Storage 多 Bucket 存储器
//
// 并发模型：
//   - bucket 切片仅通过 expansion() 在 mu.Lock 下追加，已有索引永不移动或释放
//   - 读路径（Share/Get/Size/Free/Range）不加锁
//   - Size()/Free() 通过原子计数器实现 O(1)
//   - New() 通过原子读 bucket.size 跳过已满桶，避免无效加锁
//
// Token 格式：28 个 hex 字符 = bucket(2B) + slot(4B) + random(8B)
// 解析只需查表提取前 12 个 hex 字符，零分配
type Storage struct {
	cap int
	// buckets 以原子快照发布:expansion 写时复制整体替换,
	// 读路径(Share/Get/Range/New/Delete)无锁读快照,与运行期扩容并发安全
	buckets   atomic.Pointer[[]*Bucket]
	NewSetter NewSetter
	mu        sync.Mutex   // 仅用于 expansion
	totalSize atomic.Int64 // 所有桶已占用槽位总数
	totalCap  atomic.Int64 // 所有桶容量总和
}

// load 返回当前桶快照
func (this *Storage) load() []*Bucket {
	return *this.buckets.Load()
}

// Share 从 token 前 4 个 hex 字符解析桶索引，零分配
func (this *Storage) Share(id string) (int, error) {
	bucket, ok := tokenDecodeBucket(id)
	if !ok {
		return 0, errors.New("invalid token")
	}
	if bucket < 0 || bucket >= len(this.load()) {
		return 0, errors.New("bucket index out of range")
	}
	return bucket, nil
}

// Get 按 token 获取对象，O(1) 零分配
func (this *Storage) Get(id string) (Setter, bool) {
	share, err := this.Share(id)
	if err != nil {
		return nil, false
	}
	return this.load()[share].Get(id)
}

// Size 当前已占用总数，O(1) 原子读
func (this *Storage) Size() int {
	return int(this.totalSize.Load())
}

// Free 当前空闲总数，O(1) 原子读
func (this *Storage) Free() int {
	return int(this.totalCap.Load() - this.totalSize.Load())
}

// Range 遍历所有对象
// 回调返回 false 时提前终止。禁止在回调中调用 New/Delete
func (this *Storage) Range(f func(Setter) bool) bool {
	for _, bucket := range this.load() {
		if !bucket.Range(f) {
			return false
		}
	}
	return true
}

// New 分配一个新对象
// 通过原子读 bucket.size 跳过已满桶，避免无效加锁
func (this *Storage) New(v any) Setter {
	cap32 := int32(this.cap)
	for _, bucket := range this.load() {
		if bucket.size.Load() >= cap32 {
			continue
		}
		if r := bucket.New(v); r != nil {
			this.totalSize.Add(1)
			return r
		}
	}
	return this.expansion(v)
}

// expansion 扩容：创建新桶并分配对象
func (this *Storage) expansion(v any) Setter {
	this.mu.Lock()
	defer this.mu.Unlock()
	cap32 := int32(this.cap)
	buckets := this.load()
	for _, bucket := range buckets {
		if bucket.size.Load() >= cap32 {
			continue
		}
		if r := bucket.New(v); r != nil {
			this.totalSize.Add(1)
			return r
		}
	}
	//写时复制:整体替换快照,与无锁读路径并发安全
	bucket := NewBucket(len(buckets), this.cap)
	bucket.NewSetter = this.NewSetter
	r := bucket.New(v)
	nb := make([]*Bucket, len(buckets)+1)
	copy(nb, buckets)
	nb[len(buckets)] = bucket
	this.buckets.Store(&nb)
	this.totalCap.Add(int64(this.cap))
	this.totalSize.Add(1)
	return r
}

// Delete 删除并返回已删除的对象
func (this *Storage) Delete(id string) Setter {
	share, err := this.Share(id)
	if err != nil {
		return nil
	}
	bucket := this.load()[share]
	r := bucket.Delete(id)
	if r != nil {
		this.totalSize.Add(-1)
	}
	return r
}

// Remove 批量删除，返回所有成功删除的对象
func (this *Storage) Remove(id []string) (r []Setter) {
	for _, v := range id {
		if s := this.Delete(v); s != nil {
			r = append(r, s)
		}
	}
	return r
}
