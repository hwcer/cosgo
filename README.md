# cosgo

cosgo 是一个 Go 语言的脚手架工具集合，提供了丰富的功能模块，用于快速构建和开发 Go 应用程序。

## 功能特性

- **模块管理系统**：基于生命周期的模块管理，支持初始化、启动和关闭等操作
- **配置系统**：基于 viper 实现的灵活配置管理，支持多种配置源和格式
- **事件系统**：用于模块生命周期管理和系统事件处理
- **进程管理**：支持 PID 文件、优雅关闭等功能
- **丰富的工具模块**：提供了多种实用工具，满足不同场景的需求

## 目录结构

```
cosgo/
├── await/         # 异步调用和等待机制
├── binder/        # 数据绑定（JSON、XML、YAML、表单等）
├── random/        # 随机数生成工具
├── redis/         # Redis 客户端封装
├── registry/      # 服务注册与发现
├── request/       # HTTP 请求客户端
├── safety/        # 安全相关工具
├── scc/           # 进程控制和协程管理
├── schema/        # 数据结构定义和解析
├── session/       # 会话管理
├── slice/         # 切片操作工具
├── times/         # 时间相关工具
├── utils/         # 通用工具函数
├── uuid/          # UUID 生成工具
├── values/        # 值操作工具
├── zset/          # 有序集合实现
├── config.go      # 配置系统
├── cosgo.go       # 核心功能
├── events.go      # 事件系统
├── funcs.go       # 通用函数
├── helps.go       # 帮助函数
├── init.go        # 初始化
├── module.go      # 模块接口
├── options.go     # 配置选项
├── pidfile.go     # PID 文件管理
├── pprof.go       # 性能分析
├── reload.go      # 重载功能
├── signal.go      # 信号处理
└── README.md      # 项目文档
```

## 核心模块说明

### await
提供异步调用和等待机制，支持超时控制和错误处理。适用于需要异步执行但又需要等待结果的场景。

### binder
支持多种数据格式的绑定，包括 JSON、XML、YAML、表单等。简化数据解析和验证过程。

### random
提供高质量的随机数生成功能，支持各种类型的随机值生成。

### redis
Redis 客户端封装，提供简洁的 API 接口，支持连接池管理和常用操作。

### registry
服务注册与发现模块，支持服务的注册、发现和健康检查。

### request
HTTP 请求客户端，支持各种 HTTP 方法、请求头设置、超时控制等。

### safety
安全相关工具，提供加密、解密、哈希等功能。

### scc
进程控制和协程管理，支持优雅启动和关闭、协程池管理等。

### schema
数据结构定义和解析，支持结构体标签解析、字段验证等。

### session
会话管理，支持内存存储和 Redis 存储，提供会话创建、验证和管理功能。

### slice
切片操作工具，提供各种切片操作函数，如合并、去重、查找等。

### times
时间相关工具，支持定时器、过期时间管理、时间格式化等。

### utils
通用工具函数，提供各种实用功能，如地址处理、字节操作、加密等。

### uuid
UUID 生成工具，支持多种 UUID 版本的生成。

### values
值操作工具，提供各种值类型的操作和转换功能。

### zset
有序集合实现，基于跳表算法，提供高效的有序数据结构。

## 快速开始

### 安装

```bash
go get github.com/hwcer/cosgo
```

### 基本使用

```go
package main

import (
    "github.com/hwcer/cosgo"
    "github.com/hwcer/logger"
)

// 自定义模块
type MyModule struct{}

func (m *MyModule) Id() string {
    return "mymodule"
}

func (m *MyModule) Init() error {
    logger.Trace("MyModule initialized")
    return nil
}

func (m *MyModule) Start() error {
    logger.Trace("MyModule started")
    return nil
}

func (m *MyModule) Close() error {
    logger.Trace("MyModule closed")
    return nil
}

func main() {
    // 注册模块
    cosgo.Use(&MyModule{})
    
    // 启动应用
    cosgo.Start(true)
}
```

### 配置管理

```go
// 设置默认配置
cosgo.Config.SetDefault("server.port", 8080)
cosgo.Config.SetDefault("server.host", "localhost")

// 读取配置
port := cosgo.Config.GetInt("server.port")
host := cosgo.Config.GetString("server.host")
```

### 事件处理

```go
// 注册事件处理函数
cosgo.On(cosgo.EventTypStarted, func() error {
    logger.Trace("Application started")
    return nil
})
```

## 高级特性

### 模块生命周期

每个模块都需要实现 `Module` 接口，包含以下方法：
- `Id()`: 返回模块唯一标识
- `Init()`: 模块初始化，在应用启动时调用
- `Start()`: 模块启动，在初始化完成后调用
- `Close()`: 模块关闭，在应用关闭时调用

### 配置优先级

配置值读取优先级从高到低：
1. 程序内设置（使用 `Config.Set()`）
2. 命令行参数
3. 环境变量
4. 配置文件
5. 默认值

### 优雅关闭

应用支持优雅关闭，在接收到终止信号时，会按照以下顺序处理：
1. 发出 `EventTypClosing` 事件
2. 按照相反的初始化顺序关闭各个模块
3. 等待所有模块关闭完成
4. 发出 `EventTypStopped` 事件

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 许可证

MIT