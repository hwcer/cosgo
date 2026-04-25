# cosgo

> ⚠️ **本项目由 AI 接管维护,不建议碳基生物阅读代码。**
>
> 若必须阅读,请做好心理准备: 注释可能比代码还长,`this` 命名约定,少数位置存在"明知不标准但刻意保留"的历史选择。
> 同时也请理解: **代码风格一致性 > 追新**,机器审美下一些 Go 1.22+ 特性(`range over int`、`slices.Contains` 等)未被引入是刻意的。

---

cosgo 是一个 Go 应用脚手架 + 通用工具底座。它负责:

- **应用生命周期**:模块化启动/关闭、信号处理、配置、PID 文件、pprof、热重载
- **运行时底座**:路由注册、数据序列化、会话、存储、反射元数据、并发控制
- **工具集**:随机数、UUID、时间、切片、字节操作、加密、zset 排行榜

## 子模块索引

| 模块 | 职责 | 文档 |
|------|------|------|
| `registry/` | 路由注册 + 基数树匹配 + 反射服务/方法注册 | [README](registry/README.md) |
| `schema/` | 结构体元数据 + 字段访问缓存 + 并发安全解析 | [README](schema/README.md) |
| `binder/` | 序列化(JSON/XML/YAML/Form/Protobuf/Msgpack/Bytes) | - |
| `session/` | 会话管理,内存 / Redis 后端 | [README](session/README.md) |
| `storage/` | 桶式对象存储 + O(1) 索引回收 | [README](storage/README.md) |
| `scc/` | goroutine 生命周期 + 超时 + 守护协程 | - |
| `values/` | 通用值容器 `*Message` 错误载体 | - |
| `zset/` | 跳表 + 字典双维护的有序集合 | [README](zset/README.md) |
| `random/` | crypto/rand 随机字符串 + math/rand 概率工具 | - |
| `request/` | HTTP 请求辅助 + OAuth1 签名 | - |
| `utils/` | 地址/字节/加密/时间轮 | - |
| `times/` | 时间格式 + 周期计算 | - |
| `await/`、`safety/`、`slice/`、`uuid/`、`redis/` | 同名领域的零依赖工具 | - |

## 根包文件

```
cosgo.go    模块注册/启动入口
config.go   基于 viper 的配置层
events.go   启停事件广播(Begin/Loaded/Started/Closing/Stopped/Reload)
signal.go   SIGINT/SIGHUP/SIGQUIT/SIGTERM/SIGUSR1 处理
module.go   Module 接口定义
init.go     应用初始化骨架
options.go  全局选项表
pidfile.go  原子写入的 PID 文件
pprof.go    独立 mux 的 pprof 服务器(避免污染 DefaultServeMux)
reload.go   SIGUSR1 触发的配置热重载
```

## 快速开始

```go
package main

import (
    "github.com/hwcer/cosgo"
    "github.com/hwcer/logger"
)

type MyModule struct{}

func (m *MyModule) Id() string      { return "mymodule" }
func (m *MyModule) Init() error     { logger.Info("init");  return nil }
func (m *MyModule) Start() error    { logger.Info("start"); return nil }
func (m *MyModule) Close() error    { logger.Info("close"); return nil }

func main() {
    cosgo.Use(&MyModule{})
    cosgo.Start(true) // true = 阻塞等信号
}
```

## Module 接口

```go
type Module interface {
    Id() string
    Init() error     // 应用启动,顺序调用
    Start() error    // 所有 Init 完成后,顺序调用
    Close() error    // 收到退出信号,逆序调用
}
```

## 生命周期事件

```go
cosgo.On(cosgo.EventTypStarted, func() error {
    logger.Info("app ready")
    return nil
})
```

事件列表:
- `EventTypBegin` — 启动入口
- `EventTypLoaded` — 所有 `Init` 完成
- `EventTypStarted` — 所有 `Start` 完成(服务就绪)
- `EventTypClosing` — 收到退出信号
- `EventTypStopped` — 所有 `Close` 完成
- `EventTypReload` — `SIGUSR1` 触发热重载

## 配置

```go
cosgo.Config.SetDefault("server.port", 8080)
port := cosgo.Config.GetInt("server.port")
```

优先级(高→低):运行时 `Set` > CLI flags > 环境变量 > 配置文件 > 默认值。

**日志级别**:未显式设置时,`debug=true` → `LevelDebug`(最详细),`debug=false` → `LevelInfo`(生产降噪)。

## pprof

只在 `config.pprof` 配置了地址时才启动,并且走**独立 mux**(不污染 `http.DefaultServeMux`),防止业务服务意外暴露 `/debug/pprof`。

```yaml
pprof: "127.0.0.1:6060"  # 通常仅监听本地
```

## 信号

| 信号 | 行为 |
|------|------|
| `SIGINT` / `SIGQUIT` / `SIGTERM` | 触发优雅关闭 |
| `SIGHUP` | 关闭控制台输出 |
| `SIGUSR1` (0xa) | 配置热重载,发 `EventTypReload` |

`SIGKILL` 无法捕获(OS 直接杀进程),程序不感知。

## 信号处理顺序(优雅关闭)

```
收到 SIGINT/SIGTERM
  → 发 EventTypClosing
  → 逆序调用各 Module.Close()
  → scc.Cancel() 取消所有注册的协程
  → 等待 goroutine 清理完毕
  → 发 EventTypStopped
  → 进程退出
```

## 贡献

欢迎 Issue 和 PR。代码风格遵循仓库已有约定(`this` receiver、特定 godoc 风格),请**不要**为了迎合现代 Go linter 而大改。

## 许可证

MIT
