# Session Storage 模块

## 概述

本模块为 CosGo 框架提供高性能、线程安全的会话存储实现。它采用基于桶（bucket）的设计，通过高效的索引管理来处理并发会话操作。

## 功能特性

- **线程安全**：使用适当的锁机制确保并发安全
- **高性能**：优化的索引管理，获取和释放索引的操作时间复杂度为 O(1)
- **自动扩展**：当现有桶满时自动创建新桶
- **可自定义**：支持为不同数据类型自定义 setter 实现
- **内存使用高效**：重用索引以最小化内存开销

## 安装

```bash
go get github.com/hwcer/cosgo/session/storage
```

## 使用方法

### 基本使用

```go
import (
    "github.com/hwcer/cosgo/session/storage"
)

// 创建一个容量为 1000 的新存储
store := storage.New(1000)

// 创建一个新会话
setter := store.Create("数据内容")

// 获取会话数据
s, found := store.Get(setter.Id())
if found {
    fmt.Println(s.Get())
}

// 更新会话数据
store.Set(setter.Id(), "更新的数据")

// 删除会话
store.Delete(setter.Id())
```

### 自定义 Setter

```go
// 定义自定义 setter
customSetter := func(id string, data interface{}) storage.Setter {
    return &MyCustomSetter{
        id: id,
        data: data,
        // 其他字段
    }
}

// 使用自定义 setter 创建存储
store := storage.New(1000, customSetter)
```

## 技术实现

### 核心组件

1. **Storage**：主存储管理器，处理多个桶
2. **Bucket**：一维数组存储，带索引管理
3. **Dirty**：管理空闲索引以实现高效重用
4. **Setter**：会话数据接口，包含 Id、Get、Set 方法

### 关键方法

#### Storage
- `New(cap int, creator ...NewSetter) *Storage`：创建新的存储实例
- `Get(id string) (Setter, bool)`：通过 ID 获取会话数据
- `Set(id string, v any) bool`：更新会话数据
- `Create(v any) Setter`：创建新会话
- `Delete(id string) Setter`：通过 ID 删除会话
- `Size() int`：获取当前会话数量
- `Free() int`：获取空闲容量

#### Bucket
- `NewBucket(id int, cap int) *Bucket`：创建新桶
- `Get(id string) (Setter, bool)`：通过 ID 获取会话数据
- `Set(id string, v any) bool`：更新会话数据
- `Create(v any) Setter`：在桶中创建新会话
- `Delete(id string) Setter`：通过 ID 删除会话

### 并发安全策略

- **Storage**：使用 `sync.RWMutex` 保护桶操作
- **Bucket**：使用 `sync.RWMutex` 保护写操作
- **Dirty**：无内部锁，依赖 Bucket 锁的保护
- **Set 操作**：对 Bucket 来说是读操作，业务级并发由应用逻辑处理

### 性能优化

1. **索引管理**：使用切片操作实现 O(1) 的索引获取和释放
2. **锁粒度**：最小化锁范围以减少竞争
3. **避免递归**：使用循环代替递归以防止栈溢出
4. **资源重用**：高效重用索引以最小化内存开销

## 设计原则

1. **职责分离**：存储层处理数据结构并发安全，业务层处理数据内容并发安全
2. **高效资源管理**：重用索引以最小化内存使用
3. **可扩展性**：自动扩展设计以处理增长的会话数量
4. **可靠性**：设置最大重试次数以防止无限循环

## 错误处理

- **无效 ID**：对无效会话 ID 返回 nil 或 false
- **容量问题**：当无可用索引时返回 nil
- **并发冲突**：使用锁防止冲突

## 测试

```bash
go test -v
```

## 基准测试

```bash
go test -bench .
```

## 许可证

MIT 许可证
