# cosgo/session 模块

## 功能介绍

cosgo/session 是一个功能完善的会话管理模块，主要用于处理用户登录状态和会话数据管理。该模块支持多种存储后端（内存和Redis），提供了会话创建、验证、更新、删除等核心功能，并内置了心跳机制以自动清理过期会话。

## 核心特性

- **多种存储后端**：支持内存存储和Redis存储，满足不同场景的需求
- **并发安全**：采用 Copy-on-Write 模式和互斥锁结合的方式，确保并发安全的同时保持高性能
- **无锁读操作**：读操作（如 Get、GetString 等）不需要加锁，保持高性能
- **Token 管理**：支持通过 Token 和 Refresh 方法管理会话Token
- **过期时间管理**：支持会话过期，减少安全风险
- **心跳机制**：支持自动清理过期会话
- **可扩展性**：支持自定义存储后端，只需实现 Storage 接口

## 安装

```bash
go get github.com/hwcer/cosgo/session
```

## 使用方法

### 初始化存储

#### 内存存储

```go
import "github.com/hwcer/cosgo/session"

// 初始化内存存储，设置容量为10000
session.Options.Storage = session.NewMemory(10000)
```

#### Redis存储

```go
import "github.com/hwcer/cosgo/session"

// 初始化Redis存储，设置地址和前缀
redisStorage, err := session.NewRedis("localhost:6379", "app")
if err != nil {
    // 处理错误
}
session.Options.Storage = redisStorage
```

### 创建会话

```go
import "github.com/hwcer/cosgo/session"

// 创建会话
s := session.New()
token, err := s.Create("user123", map[string]any{
    "username": "test",
    "role": "admin",
})
if err != nil {
    // 处理错误
}
// 存储token到cookie或返回给客户端
```

### 验证会话

```go
import "github.com/hwcer/cosgo/session"

// 验证会话
s := session.New()
err := s.Verify(token)
if err != nil {
    // 处理错误，例如重定向到登录页
}
// 会话验证成功，获取会话数据
username := s.GetString("username")
role := s.GetString("role")
```

### 更新会话

```go
import "github.com/hwcer/cosgo/session"

// 更新会话数据
s.Set("username", "newUsername")
s.Set("lastLogin", time.Now())
// 会话结束时会自动保存更新
```

### 删除会话

```go
import "github.com/hwcer/cosgo/session"

// 删除会话
err := s.Delete()
if err != nil {
    // 处理错误
}
// 会话已删除，用户需要重新登录
```

## 配置选项

```go
var Options = struct {
    Name string //session cookie name
    MaxAge    int64  //有效期(S)
    Secret    string //16位秘钥
    Storage   Storage
    Heartbeat int32 //心跳(S)
}{
    Name:      "_cookie_vars",
    MaxAge:    3600,
    Secret:    "UVFGHIJABCopqDNO", //redis 存储时生成TOKEN的密钥
    Heartbeat: 10,
}
```

## 核心结构

### Data 结构

```go
type Data struct {
    id        string        // 默认uuid,memory模式会定制此ID
    uuid      string        //GUID
    index     int32         //socket server id
    values    values.Values // 私有字段，外部通过方法访问
    heartbeat int32
    mutex     sync.Mutex
}
```

### Session 结构

```go
type Session struct {
    *Data
    dirty []string
}
```

### Storage 接口

```go
type Storage interface {
    New(data *Data) error                                             //同Create
    Get(id string) (data *Data, err error)                            //验证TOKEN信息
    Create(uuid string, value map[string]any) (data *Data, err error) //用户登录创建新session
    Update(data *Data, value map[string]any) error                    //更新session数据
    Delete(data *Data) error                                          //退出登录删除SESSION
}
```

## 注意事项

1. **Session 绑定请求上下文**：一个 Session 绑定的是一个用户的单次请求的上下文，不会存在并发问题
2. **业务层面限制**：业务层面会限制用户的并发请求以保证数据安全
3. **Data 结构中的 values 字段**：是私有的，外部只能通过提供的方法访问
4. **读操作无锁**：读操作（如 Get、GetString 等）是无锁的，直接访问 values 字段
5. **写操作加锁**：写操作（如 Set、Update、Delete 等）使用互斥锁保护，并采用 Copy-on-Write 模式
6. **Redis 存储过期时间**：Redis 存储实现中，会话数据会设置过期时间，避免内存泄漏
7. **内存存储心跳机制**：内存存储实现中，内置心跳机制，自动清理过期会话

## 性能优化

1. **选择合适的存储后端**：单机应用建议使用内存存储，分布式应用建议使用Redis存储
2. **合理设置过期时间**：根据业务需求合理设置会话过期时间，避免会话数据占用过多内存
3. **减少写操作**：尽量减少会话数据的写操作，因为写操作会克隆 values，产生内存开销
4. **批量更新**：使用 Update 方法批量更新会话数据，减少写操作次数

## 常见问题

### 会话验证失败

- **原因**：Token 格式不正确、Token 已过期、Token 已被替换
- **解决方法**：检查 Token 格式是否正确，重新登录获取新 Token

### 会话数据丢失

- **原因**：会话已过期、存储后端故障、写操作失败
- **解决方法**：检查存储后端状态，合理设置过期时间，处理写操作错误

### Redis 存储连接失败

- **原因**：Redis 服务未启动、网络连接失败、认证失败
- **解决方法**：检查 Redis 服务状态，检查网络连接，检查认证信息

## 代码结构

```
session/
├── data.go        // 定义会话数据结构和相关方法
├── session.go     // 实现会话管理核心功能
├── storage.go     // 定义存储接口
├── memory.go      // 实现内存存储
├── redis.go       // 实现Redis存储
├── options.go     // 定义配置选项
├── heartbeat.go   // 实现心跳机制
├── errors.go      // 定义错误常量
├── memory_setter.go // 实现内存存储的 setter
└── README.md      // 说明文档
```

## 总结

cosgo/session 模块是一个设计合理、功能完善的会话管理模块，具有并发安全、存储灵活、功能完善、性能优异、安全性高等特点。通过该模块，开发者可以更专注于业务逻辑的实现，而不用关心会话管理的细节。
