# cosgo/session

HTTP 会话管理。支持内存和 Redis 两种后端,内建心跳淘汰机制。

## 使用场景

一个 Session **绑定到一次请求上下文**,不跨请求共享,也不假设跨 goroutine 并发访问。业务层需保证同一个 session 不会被多个请求同时操作(典型 HTTP 中间件模型)。

## 存储后端选择

```go
// 内存后端(单机,无持久化)
session.Options.Storage = session.NewMemory(10000)  // 容量

// Redis 后端(分布式)
redisStorage, err := session.NewRedis("localhost:6379", "app")
if err != nil { log.Fatal(err) }
session.Options.Storage = redisStorage
```

## 核心流程

### 登录/创建

```go
s := session.New()
token, err := s.Create("u-123", map[string]any{
    "username": "alice",
    "role":     "admin",
})
if err != nil { /* ... */ }
// 把 token 放 Cookie 返回给客户端
```

### 验证

```go
s := session.New()
if err := s.Verify(token); err != nil {
    // ErrorSessionEmpty / Illegal / NotExist / Replaced
    return err
}
role := s.GetString("role")
```

Verify 的 secret 比较走零分配的常量时间字符串比较（`constantTimeStringEqual`），防时序侧信道，无 `[]byte` 转换开销。

### 写入

```go
s.Set("last_login", time.Now())
s.Update(map[string]any{"a": 1, "b": 2})
```

写入会通过 `markDirty` 累计脏 key,**请求结束时**(`s.Release()`)统一 flush 到后端存储。

### 删除

```go
if err := s.Delete(); err != nil { /* ... */ }
```

### 释放(必须)

```go
defer s.Release()  // 常在 HTTP 中间件中以 defer 形式调用
```

Release 做两件事:
1. 把累计的 dirty key flush 到 Storage.Update
2. 清空 session 的内部引用,让 Data 可以被 GC

**忘记调用会导致修改丢失。**

## 配置

```go
session.Options.Name      = "_my_sid"    // Cookie 名称
session.Options.MaxAge    = 3600         // 秒，0 = 不过期
session.Options.Heartbeat = 10           // 内存后端淘汰心跳周期（秒）
```

## Storage 接口

```go
type Storage interface {
    New(data *Data) error
    Get(id string) (*Data, error)
    Create(uuid string, value map[string]any) (*Data, error)
    Update(data *Data, value map[string]any) error
    Delete(data *Data) error
}
```

实现自定义后端实现此接口,赋值给 `session.Options.Storage` 即可。

## 事件钩子

```go
session.On(session.EventSessionNew,     func(v any) { /* *Data */ })
session.On(session.EventSessionCreated, func(v any) { /* *Data */ })
session.On(session.EventSessionRelease, func(v any) { /* *Data */ })
session.On(session.EventHeartbeat,      func(v any) { /* int32 */ })
```

事件系统走 copy-on-write + `atomic.Pointer`:
- 写路径(On):拷贝整张 listener map,原子发布新版本
- 读路径(Emit):一次 atomic load,零锁,纳秒级

详见 `events.go`。

## 并发语义

- **Data 结构**:读操作(`Get` / `GetString` / `Range` 等)无锁直接访问 `values` 字段;写操作(`Set` / `Update` / `Delete`)`sync.Mutex` + Copy-on-Write,旧快照的读者不受影响。
- **Session 结构**:绑定单次请求,调用方不应跨 goroutine 共享。
- **Data.heartbeat**:心跳协程每 N 秒 `+= N`,请求协程随时归零。故意不使用 atomic——x86-64 上 aligned int32 读写是单指令完成的,最坏丢失一次累加/归零（差 ±1 个心跳周期）,对粗粒度掉线判断无实际影响。
- **Data.Atomic()**:自增包序列号,`atomic.AddInt32` + 双 check 重置,越界安全。

## 心跳淘汰(内存后端)

`Options.Heartbeat > 0 && Options.MaxAge > 0` 时启用。

```
每 Heartbeat 秒触发一次:
  1. Range 所有 Data,累加 heartbeat 值
  2. heartbeat >= MaxAge 的收集到本地 slice
  3. Range 返回后,逐个 Storage.Delete + Emit(EventSessionRelease)
```

**关键**:先 Range 收集再 Delete,避免迭代中修改底层 storage。

## Redis 后端

- 键格式:`<prefix>-<uuid>`(默认 prefix `"cookie"`)
- 数据以 Redis Hash 存储,HMSet 写入,HGetAll 读取
- 过期时间由 `Options.MaxAge` 控制,Create/Get(续约)/Update 时都会刷 `Expire`
- `Expire` 调用失败会 `logger.Alert` 记录(不会静默吞掉)

## 安全

| 项 | 处理 |
|---|------|
| Token 比较 | `constantTimeStringEqual` 零分配常量时间比较，防时序侧信道 |
| 会话 ID 随机源 | 内存后端: `crypto/rand` 生成 28 字符 hex token（bucket+slot+8 字节随机，详见 `cosgo/storage`）；Redis 后端: 用户 UUID |
| Session 劫持防护 | Token = `secret(6 字符随机) + sessionID`，secret 存在 Data 内部，其它设备登录后 secret 刷新，旧 token 校验返回 `ErrorSessionReplaced` |

## 常见问题

| 现象 | 原因 | 处理 |
|------|------|------|
| Verify 返回 `ErrorSessionReplaced` | 其它设备/浏览器登录生成新 token 覆盖了旧 session | 引导客户端重新登录 |
| Session 数据丢失 | 请求结束未调用 `Release()` / Redis Expire 失败 / 超过 MaxAge | 检查 defer / 观察 `logger.Alert` 日志 |
| Redis 连接失败 | Redis 未启动 / 网络 / 密码 | 检查 NewRedis 返回的 error 并重试 |

## 目录结构

```
session/
├── session.go        Session 入口与生命周期
├── data.go           Data 结构与读写操作
├── data_setter.go    Setter 模板
├── data_getter.go    Getter 模板
├── storage.go        Storage 接口
├── memory.go         内存后端 + 心跳
├── memory_setter.go  内存后端的 Setter 适配
├── redis.go          Redis 后端
├── heartbeat.go      心跳定时器
├── events.go         事件订阅(CoW + atomic.Pointer)
├── options.go        全局配置
├── errors.go         错误常量
└── README.md
```
