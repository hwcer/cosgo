# cosgo/registry

路由注册 + 树匹配 + 反射服务/方法注册。cosweb 的路由层底座。

## 组成

```
Registry        顶层：按名字索引 Service
  └─ Router     路由匹配引擎
      ├─ static  静态路径表，O(1) hash 命中
      └─ radix   路由树，处理含 :param / *wild 的动态路径
  └─ Service    一批注册在同一前缀下的 Node
      └─ Node   反射 Value + 可选 Binder（method 的 receiver）
```

## 注册

### 函数

```go
r := registry.New()
svc := r.Service("api", myHandler)
svc.SetMethods([]string{"GET", "POST"})
_ = svc.Register(someFunc, "/users/:id")
```

### struct 批量注册

```go
type Users struct{}
func (Users) Get(c *cosweb.Context) any { ... }
func (Users) List(c *cosweb.Context) any { ... }

svc := r.Service("api", handler)
_ = svc.Register(&Users{})
// 得到: /api/users/get, /api/users/list
```

## 路径匹配

```go
node, params := r.Search("GET", "/api/users/42")
if node != nil {
    id := params.ByName("id")  // "42"
}
```

### 优先级

1. **静态路径**（`/api/users`）—— `Router.static` 纯 hash 查询
2. **参数路径**（`/api/users/:id`）—— 路由树按段回溯匹配
3. **通配符**（`/assets/*path`）—— 匹配剩余整段，零拷贝切片原路径

### 路径占位符

| 语法 | 含义 | 捕获 key |
|------|------|----------|
| `:name` | 单段参数 | `name` |
| `*name` | 通配符，匹配剩余整条路径 | `*`（固定） |

## Params 类型

`Search` 返回 `Params`（`[]Param` 切片），不是 `map[string]string`：

```go
type Param struct {
    Key   string
    Value string
}
type Params []Param

params.Get("id")       // (string, bool)
params.ByName("id")    // string（未找到返回 ""）
```

路由参数通常 1~3 个，切片线性查找比 map hash 更快，内存连续且缓存友好。

## 性能

### 实测（AMD Ryzen AI 9 HX 370）

| 场景 | ns/op | allocs | B/op | 对比 Gin |
|------|-------|--------|------|---------|
| **静态命中** `/api/users` | **16** | **0** | **0** | ~15-20ns |
| **参数命中** `/api/users/:id` | **55** | **1** | **32** | ~50-80ns |
| **通配符命中** `/assets/*` | **50** | **1** | **32** | ~60-100ns |
| **全 miss** `/nope/not/here` | **24** | **0** | **0** | ~30-50ns |
| `Join` 规范化 | 5.2 | 0 | 0 | — |

### 关键优化

1. **静态路径走 hash 不进树**：绝大多数 REST API 是静态路径，零分配、lock-free。
2. **零 Split 匹配**：`matchScan` 直接扫描 path 字符串提取段，不预分配 `[]string` 切片。
3. **children 拆分字段**：`statics []staticChild` + `paramChild *RadixNode` + `wildChild *RadixNode`，按优先级直接访问，无 map hash 开销。
4. **method 切片化**：`[]methodEntry` 替代 `map[string]*Node`，1~3 个方法时线性扫描更快。
5. **Params 切片**：`[]Param` 替代 `map[string]string`，回溯 O(1) reslice（不是 O(k) map delete）。
6. **toLowerFast**：匹配时先检查段是否已全小写，是则跳过 `Formatter` 函数调用。
7. **通配符后缀零拷贝**：`path[pos:]` 直接切原字符串。
8. **params 懒分配**：无参路径返回 nil Params，零分配。

## 并发模型

- **注册期**（`Service` / `Register`）约定在启动阶段单 goroutine 内完成。内部不加锁。
- **请求期**（`Search`）纯读，lock-free。
- 运行时动态注册需调用方自行同步。
- `Params` 切片由 `Match` 每次调用独立创建，在调用栈内独享，不跨 goroutine。

## 路由字符串约定

`Formatter = strings.ToLower`。所有静态段 lowercase 后匹配。

## API

```go
// Registry
r := registry.New()
r.Service(name string, hs ...Handler) *Service
r.Get(name string) (*Service, bool)
r.Search(method string, paths ...string) (*Node, Params)
r.Range(func(*Service) bool)
r.Nodes(func(*Node) bool)

// Service
svc.Register(i any, prefix ...string) error
svc.RegisterWithMethod(i any, methods []string, prefix ...string) error
svc.Paths() []string
svc.Range(func(*Node) bool)

// Router
router.Search(method string, paths ...string) (*Node, Params)
router.Register(node *Node, methods []string) error

// Node
node.Name() string
node.IsFunc() / IsMethod() / IsStruct() bool
node.Call(args ...any) []reflect.Value
node.Value() / Binder() / Method()

// Params
params.Get(name string) (string, bool)
params.ByName(name string) string
```

## format() 占位符

| 占位符 | 值 |
|--------|-----|
| `%v` | `funcName` 或 `structName/methodName` |
| `%s` | `structName`（仅 struct 路径） |
| `%m` | `funcName` 或 `methodName` |

## 目录结构

```
registry/
├── registry.go               Registry 入口 + Service 管理
├── router.go                 Router + RadixNode 路由树匹配
├── params.go                 Param/Params 切片类型
├── node.go                   Node 反射包装
├── service.go                Service 注册与解析
├── func.go                   Join/Split/Route 等工具函数
├── options.go                Formatter 全局配置
├── bench_test.go             性能基准
├── router_test.go            路由功能测试
├── route_test.go             路由匹配测试
├── method_batch_test.go      批量方法注册测试
├── method_duplicate_test.go  重复注册检测测试
├── param_multiple_test.go    多参数名测试
└── README.md
```
