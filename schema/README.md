# cosgo/schema

结构体元数据 + 字段访问缓存。被 cosmo / updater 等 ORM 层用作反射底座。

## 设计目标

1. **运行期零 lock**:所有 Parse 缓存用 `sync.Map`,读者命中快路径约 80 ns。
2. **首次并发构建 μs 级唤醒**:`initDone chan struct{}` 替代 ms 级轮询。
3. **字段访问编译化**:`getValueFn/setValueFn` 在 Schema 构建期预热,运行期是纯函数调用(非反射每次查 Index)。
4. **多别名统一查找**:Go 字段名 / db 标签 / json 标签 任一命中即返回,单次 map 查询。

## 快速开始

```go
type User struct {
    ID   int64  `json:"id"   bson:"_id"`
    Name string `json:"name" bson:"name"`
    Age  int    `json:"age"  bson:"age"`
}

// 解析(结果自动缓存)
sch, err := schema.Parse(&User{})
if err != nil {
    panic(err)
}

// 按任意别名查字段
f := sch.LookUpField("Name")   // Go 名
f := sch.LookUpField("_id")    // db 名
f := sch.LookUpField("age")    // json 名

// 读写值
u := &User{}
name := sch.GetValue(u, "Name")         // any
_ = sch.SetValue(u, "alice", "Name")    // error
```

## 启动期预热(推荐)

高并发场景下,首次访问新类型时 Parse 会导致其它请求等待构建完成。启动时调用一次预热可完全消除这个等待:

```go
func main() {
    if err := schema.Warm(
        &model.User{}, &model.Order{}, &model.Product{},
    ); err != nil {
        log.Fatal(err)
    }
    // ... 后续 cosmo/updater 不再触发同步构建等待
}

// 自定义 Options 版本
schema.WarmWithOptions(opts, &T1{}, &T2{})
```

## Schema 结构

```go
type Schema struct {
    Name           string             // Go 类型名
    Table          string             // 表名(Tabler 接口 / TableName 生成规则 / special)
    Embedded       []*Field           // 匿名嵌入字段
    ModelType      reflect.Type       // 原始 Go 类型
    Fields         map[string]*Field  // 按 Go 字段名(含嵌入提升)的全量索引,公开
    // 以下为私有实现细节:
    // unifiedFields  map[string]*Field  LookUpField 的单次查询索引(Go 名/db/json 任一)
    // dbFields       []*Field           带 db 标签字段的有序列表(Schema.Range 用)
    // initDone       chan struct{}      并发构建同步信号
}
```

### 别名优先级(LookUpField)

Go 字段名 > db 标签 > json 标签。冲突时先匹配的字段返回,后来者不覆盖。

## API 参考

### 解析

```go
Parse(dest any) (*Schema, error)
GetOrParse(dest any, opts *Options) (*Schema, error)
ParseWithSpecialTableName(dest any, tableName string, opts *Options) (*Schema, error)

Warm(dests ...any) error
WarmWithOptions(opts *Options, dests ...any) error
```

### Schema 方法

```go
sch.LookUpField(name string) *Field
sch.GetValue(obj any, key string, keys ...any) any
sch.SetValue(obj any, val any, key string, keys ...any) error
sch.Range(cb func(*Field) bool)    // 遍历带 db 标签的字段
sch.JSName(k string) string
sch.New() reflect.Value             // 创建 ModelType 的新实例
sch.Make() reflect.Value            // 创建 []*ModelType 的新切片
```

### Field 方法

```go
f.Name                    // Go 字段名
f.FieldType               // reflect.Type
f.IndirectFieldType       // 解引用后的 Type
f.Index                   // 字段路径(含嵌入提升后的完整路径)
f.StructField             // 原始 reflect.StructField
f.Schema                  // 所属 Schema
f.Embedded                // 嵌入子对象的 Schema(仅 struct 类型字段)

f.JSName() string         // json 标签(缓存)
f.DBName() string         // bson 或 json 标签(缓存)
f.GetName(tags ...) string

f.Get(v reflect.Value) reflect.Value
f.Set(v reflect.Value, val any) error
```

## 性能

### 实测基准(AMD Ryzen AI 9 HX 370)

| 操作 | ns/op | allocs/op | B/op |
|------|-------|-----------|------|
| `LookUpField("Name")` (Go 名) | **5.6** | **0** | **0** |
| `LookUpField("_id")` (db 名) | 5.6 | 0 | 0 |
| `LookUpField(miss)` | 4.0 | 0 | 0 |
| `GetValue(obj, "Name")` | 35 | 1 | 16 |
| `SetValue(obj, "alice", "Name")` | 19 | 0 | 0 |
| **`Parse` 缓存命中** | **19** | **0** | **0** |
| `ParseWithSpecialTableName` 缓存命中 | 69 | 2 | 96 |

### 优化要点

1. **字段访问函数预编译**:`buildFieldMappings` 末尾遍历 `Fields` 调用 `compileGetValueFn/compileSetValueFn`,运行期调用时是直接的 closure 调用,不走 `field.ReflectValueOf + reflect.Indirect + Field(i)` 动态流程。
2. **LookUpField 单次 map**:`unifiedFields` 在构建期按优先级合并 Go/db/json 三套别名到一张 map。
3. **GetValue/SetValue 单 key 快路径**:common case 不做 `append([]any{key}, keys...)`。
4. **Parse 缓存键 struct 化**:带 `specialTableName` 时用 `struct{t reflect.Type; name string}`,避免 `fmt.Sprintf("%p-%s", ...)` 的字符串分配。
5. **cache-hit 快路径**:先 `Store.Load`,命中立即返回;未命中才 `make(Schema{})` 并 `LoadOrStore`,消除 cache 命中路径的 Schema 分配。
6. **defer/recover 隔离**:缓存命中在 `defer` 之前返回,避免闭包逃逸导致的堆分配。`parseSchemaSlow` 独立函数承载完整解析逻辑和 panic 保护,仅缓存未命中时调用。

## 并发构建的 chan 信号机制

多 goroutine 同时 Parse 同一新类型时:

```
G1: Load miss → LoadOrStore wins → 持 Schema,开始 Init
G2: Load hit (G1 刚 Store) → waitSchemaInit(s)
G3: 同上
  ...
G1: Init 完成 → defer close(s.initDone) → 所有等待者被唤醒
G2/G3: <-s.initDone 返回 → return (s, s.err)
```

唤醒延迟是 OS 线程调度级(μs),不是轮询级(ms)。

超时兜底:`SchemaInitTimeout = 30s`,防御同 goroutine 自引用 struct 造成的死锁。可用 `schema.SchemaInitTimeout = ...` 调整。

## 嵌入字段的注意事项

匿名嵌入字段会被**提升**到外层 Schema 的 `Fields` 中,但存的是**副本**(Index 已调整为外层视角)。这些副本在构建期由 `warmupFieldFns` 按新 Index 重新编译访问函数,保证访问的是真正的嵌入字段而不是同索引的外层字段。

```go
type Inner struct { Name string }
type Outer struct {
    ID int
    Inner  // 匿名嵌入,Name 被提升
}

sch, _ := schema.Parse(&Outer{})
nameField := sch.LookUpField("Name")
// nameField.Index = [1, 0]  (Outer 的第 1 个字段 → Inner 的第 0 个字段)
// 通过 nameField.Get(outerValue) 可以正确拿到 Outer.Inner.Name
```

## 配置

```go
opts := schema.New(&NamingStrategy{TablePrefix: "t_"})
opts.Store = &sync.Map{}              // 自定义缓存(通常不需要)
sch, _ := opts.Parse(&User{})
```

默认 `config = schema.New()`,被所有无 opts 的调用共享。

## 错误

- `ErrUnsupportedDataType` — dest 不是结构体
- `ErrDuplicateDBName` — 同一 Schema 内两个字段的 db 标签冲突
- `schema parse panic: ...` — 构建期 panic 被 recover 转换为 error 返回
- `schema init timeout after 30s ...` — 自引用 struct 或 init 失败导致无限等待
