# cosgo/storage

桶式对象存储。面向**读远大于写**的高速缓存场景（如 SESSION），提供 O(1) 分配/释放 + 无锁 O(1) 按 ID 查找。

## 基本使用

```go
import "github.com/hwcer/cosgo/storage"

s := storage.New(1000)                 // 容量 1000 的初始桶
setter := s.New(myData)                // 写入，返回带 Token ID 的 Setter
got, ok := s.Get(setter.Id())          // 按 ID 取回（无锁，~8ns）
_ = s.Delete(setter.Id())              // 删除
count := s.Size()                      // O(1) 原子读
free  := s.Free()                      // O(1) 原子读
```

## 核心设计

### Token ID

对象 ID 是 28 个 hex 字符的 token，编码了**定位信息 + 随机防猜测**：

```
Token 结构（14 字节 = 28 hex 字符）：

  ┌──────────┬──────────┬──────────────────────┐
  │  bucket  │   slot   │       random         │
  │  2 bytes │  4 bytes │       8 bytes        │
  │  uint16  │  uint32  │  crypto/rand         │
  └──────────┴──────────┴──────────────────────┘

示例: "002a000000f3a7c3e19d5b02f841"
       ^^^^                          bucket=42
           ^^^^^^^^                  slot=243
                   ^^^^^^^^^^^^^^^^  8 字节随机数
```

- **定位**：Get 时解析前 12 个 hex 字符（查表运算，~3ns），得到 bucket + slot 直接数组寻址
- **防猜测**：尾部 8 字节由 `crypto/rand` 生成（2^64 种可能），暴力猜测不可行
- **防重放**：同一槽位复用后 random 不同，旧 token 的全串比较必定失败
- **可作为客户端 TOKEN**：固定 28 字符，不含结构信息泄露，不可被序号推测

### Bucket：一维数组 + 脏索引管理

```
Bucket.values []unsafe.Pointer  // 预分配固定长度，原子读写，生命周期内不扩不缩
Bucket.dirty  *dirty            // 空闲槽位 LIFO 栈，O(1) Acquire/Release，位图防双重释放
```

**值**存在 `values[index]`（通过 `atomic.LoadPointer`/`StorePointer` 原子操作），**索引**由 dirty 以 LIFO 栈方式分配和回收。

### Storage：多 Bucket 动态扩容

```
Storage.bucket []*Bucket   // append-only 增长，从不收缩
Storage.expansion()        // 所有 bucket 都满时创建新桶并追加
Storage.totalSize/totalCap // 原子计数器，O(1) 查询
```

`New()` 通过原子读 `bucket.size` 跳过已满桶，避免对满桶加锁。

## 并发模型

### Bucket 层

- `values` 切片**预分配固定长度**，生命周期内不变 → slice header 无 race
- 每个槽位通过 `atomic.LoadPointer`/`StorePointer` 原子读写 → **无 torn read**
- 写路径（`New` / `Delete`）使用 `sync.Mutex` 串行化
- 读路径（`Get` / `Range` / `Size` / `Free`）**完全无锁**：
  - 最坏情况读到 stale 值（刚被 Delete 或刚被 New），但不会 panic 或读到非法内存
  - 调用方容忍短暂不一致（如 Range 回调中发现 Setter 已失效应跳过）
- **禁止**在 `Range` 回调里调用 `New`/`Delete`

### Storage 层

- `bucket` 切片只通过 `expansion()` 在 `mu.Lock` 下 append，已有索引永不移动或释放
- 读路径不加锁：并发扩容时读侧持有旧 slice header，看到"少一个 bucket"的快照，已有 `*Bucket` 指针仍有效
- `Size()`/`Free()` 通过 `atomic.Int64` 原子计数器实现，O(1)

## 性能

| 操作 | 延迟 | 分配 | 说明 |
|------|------|------|------|
| `Get(id)` | **~8 ns** | **0** | hex 查表解析 + 原子指针读 + 全串比较 |
| `New(v)` | ~100 ns | 3 | token 字符串 + Setter 对象 + entry 包装 |
| `Delete(id)` | ~100 ns | 0 | 原子写 nil + dirty.Release |
| `Range(f)` | O(n) | 0 | 遍历所有槽位，原子读 |
| `Size()` / `Free()` | **~0.3 ns** | **0** | 单次原子读 |

随机数生成使用 256 字节缓冲池摊薄 `crypto/rand` 系统调用，每 32 次 `New()` 才触发一次。

## API

```go
// 创建
storage.New(cap int, creator ...NewSetter) *Storage

// 写入
s.New(v any) Setter                    // 分配新对象，返回带 token ID 的 Setter
s.Delete(id string) Setter             // 删除并返回对象
s.Remove(id []string) []Setter         // 批量删除

// 查询
s.Get(id string) (Setter, bool)        // 按 token 获取，O(1) 无锁
s.Size() int                           // 已占用总数，O(1) 原子读
s.Free() int                           // 空闲总数，O(1) 原子读
s.Range(f func(Setter) bool) bool      // 遍历所有对象
s.Share(id string) (int, error)        // 解析 token 中的桶索引
```

## 自定义 Setter

```go
type MySetter struct {
    id   string
    data any
}
func (s *MySetter) Id() string { return s.id }

creator := func(id string, v any) storage.Setter {
    return &MySetter{id: id, data: v}
}
s := storage.New(1000, creator)
```

## 关键约束

1. **Range 回调内不能 New/Delete**。先收集 ID，再批量处理。`session/memory.go` 的 Heartbeat 是标准示例。
2. **返回的 Setter 可能被并发 Delete**。调用方应先 `Get` 验证有效再操作。
3. **Bucket 容量固定**。只有 Storage 层会扩容（新建 Bucket 追加），旧 Bucket 永不扩也永不缩。
4. **Token 不要手工构造**。始终从 `Setter.Id()` 或 `Storage.New` 返回值获取。
5. **bucket 索引上限 65535**（uint16），slot 索引上限 ~42 亿（uint32）。

## 目录结构

```
storage/
├── storage.go         Storage 多桶容器 + 扩容
├── bucket.go          Bucket 单桶 + 原子槽位操作
├── dirty.go           空闲索引栈 + 位图防双重释放
├── token.go           Token 编解码（hex + crypto/rand）
├── setter.go          Setter 接口 + 默认实现
├── storage_test.go    功能/并发/性能测试
└── README.md
```
