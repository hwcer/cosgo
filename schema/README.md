# cosgo/schema

`cosgo/schema` 是一个用于 Go 语言的结构体反射和元数据管理库，提供了高效的字段访问、类型转换和 schema 解析功能。

## 功能特性

- **结构体反射**：自动解析结构体字段信息，支持嵌套结构体和嵌入字段
- **标签解析**：支持解析 `json`、`bson` 等标签，提取字段的元数据
- **字段访问**：提供统一的字段访问接口，支持获取和设置字段值
- **类型转换**：提供类型转换工具，支持各种类型之间的转换
- **并发安全**：使用原子操作和 sync.Map 确保并发安全
- **性能优化**：预分配内存、缓存常用操作，减少反射开销

## 安装

```bash
go get github.com/hwcer/cosgo/schema
```

## 快速开始

### 基本用法

```go
package main

import (
    "fmt"
    "github.com/hwcer/cosgo/schema"
)

// 定义结构体
type User struct {
    ID   int    `json:"id" db:"user_id"`
    Name string `json:"name" db:"user_name"`
    Age  int    `json:"age" db:"user_age"`
}

func main() {
    // 解析结构体
    user := User{ID: 1, Name: "Alice", Age: 25}
    sch, err := schema.Parse(&user)
    if err != nil {
        fmt.Println("Error parsing schema:", err)
        return
    }
    
    // 获取字段信息
    field, ok := sch.Fields["ID"]
    if ok {
        fmt.Println("Field Name:", field.Name)
        fmt.Println("JSON Name:", field.JSName())
        fmt.Println("DB Name:", field.DBName())
    }
    
    // 获取字段值
    value, err := sch.GetValue(&user, "Name")
    if err != nil {
        fmt.Println("Error getting value:", err)
    } else {
        fmt.Println("Name Value:", value)
    }
    
    // 设置字段值
    err = sch.SetValue(&user, "Age", 26)
    if err != nil {
        fmt.Println("Error setting value:", err)
    } else {
        fmt.Println("Updated Age:", user.Age)
    }
}
```

### 高级用法

#### 自定义表名

```go
// 实现 Tabler 接口
type User struct {
    ID   int    `json:"id" db:"user_id"`
    Name string `json:"name" db:"user_name"`
}

func (u User) TableName() string {
    return "custom_users"
}

// 或者使用自定义表名参数
sch, err := schema.ParseWithSpecialTableName(&User{}, "custom_users", nil)
```

#### 嵌入结构体

```go
type Base struct {
    ID        int       `json:"id" db:"id"`
    CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type User struct {
    Base
    Name string `json:"name" db:"name"`
    Age  int    `json:"age" db:"age"`
}

// 嵌入字段会被自动合并
sch, _ := schema.Parse(&User{})
fmt.Println("Has ID field:", sch.Fields["ID"] != nil)
fmt.Println("Has Name field:", sch.Fields["Name"] != nil)
```

## API 文档

### 主要函数

#### `Parse(dest interface{}) (*Schema, error)`
从目标结构体解析 Schema 信息。

- **参数**：
  - `dest`：目标结构体实例或指针
- **返回**：
  - `*Schema`：解析生成的 Schema 实例
  - `error`：解析过程中发生的错误

#### `GetOrParse(dest interface{}, opts *Options) (*Schema, error)`
从目标结构体解析 Schema 信息，支持自定义选项。

- **参数**：
  - `dest`：目标结构体实例或指针
  - `opts`：解析选项，包含缓存、表名生成等配置
- **返回**：
  - `*Schema`：解析生成的 Schema 实例
  - `error`：解析过程中发生的错误

#### `ParseWithSpecialTableName(dest interface{}, specialTableName string, opts *Options) (*Schema, error)`
从目标结构体解析 Schema 信息，支持自定义表名。

- **参数**：
  - `dest`：目标结构体实例或指针
  - `specialTableName`：自定义表名
  - `opts`：解析选项，包含缓存、表名生成等配置
- **返回**：
  - `*Schema`：解析生成的 Schema 实例
  - `error`：解析过程中发生的错误

### Schema 方法

#### `GetValue(value interface{}, path string) (interface{}, error)`
获取字段值。

#### `SetValue(value interface{}, path string, val interface{}) error`
设置字段值。

#### `Range(cb func(*Field) bool)`
遍历所有字段。

#### `New() reflect.Value`
创建一个新的结构体实例。

#### `Make() reflect.Value`
创建一个新的结构体切片。

### Field 方法

#### `Name() string`
获取字段名称。

#### `JSName() string`
获取 JSON 字段名称。

#### `DBName() string`
获取数据库字段名称。

#### `Type() reflect.Type`
获取字段类型。

#### `GetValue(value reflect.Value) reflect.Value`
获取字段的反射值。

#### `Set(value reflect.Value, val interface{}) error`
设置字段值。

## 性能优化

### 已实施的优化

1. **并发性能**：
   - 使用原子操作替代 channel 实现并发控制
   - 优化 `waitSchemaInit` 函数，减少 CPU 占用

2. **内存使用**：
   - 预分配 map 容量，减少内存分配
   - 移除不可控的缓存，避免内存泄漏
   - 使用 sync.Map 实现并发安全的缓存

3. **代码结构**：
   - 拆分解析逻辑为多个函数，提高代码可读性
   - 添加详细的 godoc 注释，提高代码可维护性
   - 统一错误处理方式，提供更清晰的错误信息

### 性能注意事项

- **反射开销**：反射操作相对较慢，但库已经自动缓存了所有 Schema 解析结果，无需手动缓存
- **内存使用**：已移除内部数值缓存，内存使用可控，无需额外监控
- **并发访问**：在高并发场景下，使用 `GetOrParse` 函数并传入自定义的 `Options` 以获得更好的性能

## 配置选项

```go
// Options 解析选项
type Options struct {
    Store    sync.Map       // 缓存存储
    TableName func(string) string // 表名生成函数
}

// 默认配置
var config = &Options{
    Store: sync.Map{},
    TableName: func(name string) string {
        return name
    },
}
```

## 错误处理

库定义了以下错误类型：

- `ErrUnsupportedDataType`：不支持的数据类型
- `ErrDuplicateFieldName`：字段名重复
- `ErrDuplicateDBName`：数据库字段名重复
- `ErrInvalidTag`：无效的标签格式
- `ErrFieldNotFound`：字段不存在
- `ErrNotObject`：不是对象类型

## 贡献指南

1. **Fork 仓库**
2. **创建分支**：`git checkout -b feature/your-feature`
3. **提交修改**：`git commit -m "Add your feature"`
4. **推送分支**：`git push origin feature/your-feature`
5. **创建 Pull Request**

## 测试

运行测试：

```bash
go test -v ./...
```

## 许可证

MIT License

## 联系方式

- **作者**：hwcer
- **邮箱**：hwcer@163.com
- **GitHub**：https://github.com/hwcer/cosgo
