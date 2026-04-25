# cosgo/safety

网络黑白名单检查器。零分配 IPv4 匹配，支持单 IP、范围、CIDR 三种规则格式。

## 快速开始

```go
import "github.com/hwcer/cosgo/safety"

s := safety.New()

// 添加规则（支持三种格式）
s.Update("cdn",    "1.2.3.4",                  safety.StatusEnable,  false)  // 单 IP
s.Update("office", "10.0.0.0~10.0.0.255",      safety.StatusEnable,  false)  // 范围
s.Update("block",  "192.168.100.0/24",          safety.StatusDisable, false)  // CIDR

// 信任所有内网 IP（127.0.0.1 + 三大私网段）
s.UseLocalAddress()

// 检查 IP
status := s.Match("10.0.0.42", true)   // StatusEnable（命中 office 规则）
status  = s.Match("8.8.8.8", false)    // StatusNone（未命中）
status  = s.Match("192.168.100.5", false) // StatusDisable（命中 block 规则）
```

## 规则格式

| 格式 | 示例 | 说明 |
|------|------|------|
| 单 IP | `"1.2.3.4"` | 精确匹配 |
| 范围 | `"10.0.0.0~10.0.0.255"` | 起止 IP（含两端） |
| CIDR | `"10.0.0.0/8"` | 子网掩码，自动计算范围 |

常见 CIDR：

| CIDR | 范围 | 地址数 |
|------|------|--------|
| `/8` | x.0.0.0 ~ x.255.255.255 | 1600 万 |
| `/12` | x.y.0.0 ~ x.y+15.255.255 | 100 万 |
| `/16` | x.y.0.0 ~ x.y.255.255 | 65536 |
| `/24` | x.y.z.0 ~ x.y.z.255 | 256 |
| `/32` | 单 IP | 1 |

## 状态

| 常量 | 值 | 含义 |
|------|---|------|
| `StatusNone` | 0 | 未匹配任何规则 |
| `StatusEnable` | 1 | 白名单（允许） |
| `StatusDisable` | 2 | 黑名单（拒绝） |

## API

```go
safety.New() *Safety

s.Update(name, rule string, status Status, local bool)  // 添加/更新规则
s.Delete(name string)                                    // 删除规则
s.Reload(f SafetyUpdate)                                 // 批量更新（回调返回 true 应用变更）
s.Get(name string) *SafetyRule                           // 按名查找规则

s.Match(ip string, useLocalAddress bool) Status          // 检查 IP（零分配，~15ns）
s.UseLocalAddress()                                      // 信任所有内网 IP
s.Lock(f func())                                         // 在 mutex 下执行自定义操作
```

### Match 参数

- `ip`：IPv4 地址字符串，自动剥离端口（`"192.168.1.2:8080"` → 只匹配 IP 部分）
- `useLocalAddress`：是否启用内网规则（`local=true` 的规则仅在此参数为 `true` 时生效）

## 并发模型

- **读路径**（Match / Get）：通过 `atomic.Pointer` 无锁读取当前规则快照，零锁争用
- **写路径**（Update / Delete / Reload）：`sync.Mutex` 串行化 + Copy-on-Write
- **内存模型**：规则指针通过 `atomic.Pointer` 发布，符合 Go 内存模型

读写不互斥：写操作复制整个规则集修改后原子替换，读操作始终看到完整一致的快照。

## 性能

| 操作 | 延迟 | 分配 | 说明 |
|------|------|------|------|
| **Match 命中** | **15 ns** | **0** | 内联 IPv4 解析 + slice 遍历 |
| Match 未命中（24 条规则） | 31 ns | **0** | 遍历全部规则无命中 |
| **并行 Match** | **1.3 ns** | **0** | 多核近线性扩展 |
| IPv4 解析 | 15 ns | **0** | 零分配内联解析 |

Match 热路径全程零堆分配：IPv4 字符串直接在栈上解析为 uint32，规则遍历使用预构建的 slice（非 map）。

## 内网规则

调用 `UseLocalAddress()` 后自动添加四条内网白名单规则：

```
127.0.0.1         localhost
10.0.0.0/8        A 类私网
172.16.0.0/12     B 类私网
192.168.0.0/16    C 类私网
```

这些规则的 `local=true`，仅在 `Match(ip, useLocalAddress=true)` 时生效。

## 目录结构

```
safety/
├── init.go           Status 常量定义
├── ipv4.go           零分配 IPv4 解析 + CIDR/范围解析
├── safety.go         Safety 主结构 + CoW 规则管理 + atomic.Pointer 无锁读
├── safety_test.go    功能测试 + 并发测试 + 性能基准
└── README.md
```
