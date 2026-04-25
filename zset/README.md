# cosgo/zset

基于**跳表 + 字典**双维护的有序集合。适合实时排行榜、Top-N 统计、优先级队列等场景。

## 特性概览

| 特性 | 说明 |
|------|------|
| **有序结构** | 跳表 O(log N) 增删查改 |
| **按键定位** | dict 辅助 O(1) 按 ID 查分数 |
| **同分保序** | 相同分数按插入顺序排列（FIFO：先到者排名靠前） |
| **守门员机制** | `maxSize` 软上限，分数未入围的新元素直接拒绝 |
| **无锁预检** | `CanEnter()` 原子读守门员分数，调用方可在业务层零锁过滤无效写入 |
| **懒淘汰** | 超过 maxSize 不立即裁剪，减少高频删除开销 |
| **线程安全** | 内置 `sync.RWMutex`，读写并发安全 |

## 快速开始

```go
// 默认降序（分数高排名靠前），不限人数
z := zset.New()

// 降序 + 限 100 人
z := zset.NewWithMaxSize(100)

// 升序（分数低排名靠前）+ 限 100 人
z := zset.NewWithMaxSize(100, 1)

z.ZAdd(1500, "alice")
z.ZAdd(1200, "bob")
z.ZAdd(1800, "carol")

rank, score := z.ZRank("carol")    // rank=0, score=1800
top10 := z.ZRange(0, 9)            // []ZNode
z.ZRem("bob")
```

---

## 核心规则

### 规则一：排序

- **降序模式**（默认，order < 0）：分数高者排名靠前——适用于主流排行榜。
- **升序模式**（order > 0）：分数低者排名靠前——适用于延迟队列/最早到期。
- **同分先到先得**：相同分数的元素按插入时间排序，先插入的排名靠前，不合并不覆盖。

跳表链表顺序与排名方向一致：表头即 rank=0。

```
示例（降序）：
  ZAdd(100, "A")   → A:100  rank=0
  ZAdd(100, "B")   → A:100  rank=0, B:100  rank=1   ← B 后到，排在 A 后面
  ZAdd(200, "C")   → C:200  rank=0, A:100  rank=1, B:100  rank=2
```

### 规则二：守门员（Gatekeeper）

当设置了 `maxSize > 0` 时，守门员机制在排行榜满员后自动启用：

- **守门员**即当前排行榜最后一名（第 maxSize 名）。
- **新成员**的分数必须**严格优于**守门员才能进入排行榜。
  - 降序模式：新分数必须 > 守门员分数。
  - 升序模式：新分数必须 < 守门员分数。
  - 同分视为**不优于**守门员，新成员被拒绝（先到者保留位置）。
- **已有成员**更新分数时不受守门员限制，始终允许更新。

```
流程：
  ZAdd(score, key):
    1. key 已存在  → 更新分数，调整排名位置
    2. 未满员      → 正常插入
    3. 已满 & 新分数优于守门员 → 插入，守门员更新
    4. 已满 & 新分数不优于守门员 → 拒绝（返回 0），不动跳表也不动 dict
```

**无锁预检**：`CanEnter(score)` 通过原子读守门员分数，不获取任何锁即可判断分数是否有可能入榜。调用方可在业务层预过滤，避免无效的 ZAdd 争抢写锁：

```go
if z.CanEnter(score) {
    z.ZAdd(score, key)
}
```

### 规则三：懒淘汰

守门员机制会产生冗余数据（跳表中超出 maxSize 的元素），采用延迟清理策略：

- **不在每次 ZAdd 后裁剪**，而是累积到阈值（`cleanupBufferSize = 100`）时批量清理。
- 清理时删除跳表末尾超出 maxSize 的所有元素，同时从 dict 中移除对应条目。
- **前 N 名始终准确**：`ZCard()`、`ZRange()`、`ZRank()` 等接口保证返回 maxSize 以内的正确数据。
- 跳表实际长度可能**暂时**超过 maxSize，属于正常现象。

这么设计是为了在高频添加场景下减少删除操作频次，以空间换吞吐。

---

## API

```go
// 创建
zset.New(order ...int8) *ZSet                          // 默认降序(order=-1)
zset.NewWithMaxSize(maxSize int32, order ...int8) *ZSet // 带人数限制

// 写入
z.ZAdd(score int64, key string) int64      // 添加或更新，返回最终分数（被拒绝返回 0）
z.ZIncr(score int64, key string) int64     // 增量更新
z.ZRem(key string) bool                    // 删除
z.CanEnter(score int64) bool               // 无锁预检：分数是否有可能入榜

// 批量删除
z.ZRemRangeByRank(start, stop int64) int64   // 按排名区间删除（0-based）
z.ZRemRangeByScore(min, max int64) int64     // 按分数区间删除

// 查询
z.ZScore(key string) (score int64, ok bool)              // O(1) 查分数
z.ZRank(key string) (rank int64, score int64)             // 查排名(0-based)+分数；不存在或未入围返回 rank=-1
z.ZElement(rank int64) (key string, score int64)          // 按排名取元素
z.ZRange(start, end int64) []ZNode                        // 按排名区间取（支持负数索引）
z.ZRangeWithCallback(start, end int64, f func(int64, string))
z.ZRangeByScore(min, max int64) []ZNode                   // 按分数区间取
z.ZRangeByScoreWithCallback(min, max int64, f func(int64, string))
z.ZCard() int64                                            // 当前有效元素个数（≤ maxSize）
z.ZCount(min, max int64) int64                            // 分数区间计数，O(log N)
```

## 分数为 0

`0` 是有效分数，与其它分数同等处理。

## 并发

`ZSet` 内部有 `sync.RWMutex`，所有公开方法均线程安全。写操作（ZAdd/ZIncr/ZRem 等）取写锁，读操作（ZRank/ZRange/ZScore 等）取读锁。

唯一的例外是 `CanEnter()`：它通过 `atomic` 读取守门员分数，**不获取任何锁**，可在高并发写入场景下做预过滤。

## 目录结构

```
zset/
├── zset.go              ZSet 主结构 + API
├── skiplist.go          跳表实现
├── zset_test.go         功能与并发测试
├── performance_test.go  性能场景测试
├── benchmark_test.go    Go Benchmark + 综合基准测试
└── README.md
```
