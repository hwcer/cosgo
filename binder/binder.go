package binder

import (
	"errors"
	"fmt"
	"io"
	"strings"
)

// binderMap 按 content-type 索引的 Binder 注册表。
//
// 并发模型(重要):
//   - 所有 Register 约定在 init 阶段完成(包级别 init 函数或 main 启动早期单 goroutine 调用),
//     之后进入只读状态。Get / Marshal / Unmarshal / Encode / Decode 都是 lock-free 纯读。
//   - 不加锁。运行时若需动态 Register,调用方必须自行加锁并承担与并发读者的竞态责任。
var binderMap = make(map[string]Binder)

type Binder interface {
	Id() uint8                           // 1
	Name() string                        //JSON
	String() string                      //application/json
	Encode(io.Writer, any) error //同Marshal
	Decode(io.Reader, any) error //同Unmarshal
	Marshal(any) ([]byte, error)
	Unmarshal([]byte, any) error
}

func New(t string) (b Binder) {
	return Get(t)
}

func ContentTypeFormat(c string) string {
	c = strings.ToLower(c)
	if i := strings.Index(c, ";"); i >= 0 {
		c = c[:i]
	}
	return strings.TrimSpace(c)
}

// Type 按 string(MIME 或 Name) / uint8(Id) 检索注册的 MIME 元数据。
// 注意:每次调用都做 map 查表。bindings 的 Id()/Name() 内部会调用本函数,
// 高频路径如 binder.Json.Id() 每次产生一次 map 查找。如果成为瓶颈,
// 调用方可把结果缓存到本地变量复用。
func Type(i any) (r *T) {
	switch v := i.(type) {
	case string:
		if strings.Contains(v, "/") {
			r = mimeTypes[ContentTypeFormat(v)]
		} else {
			r = mimeNames[strings.ToUpper(v)]
		}
	case int:
		r = mimeIds[uint8(v)]
	case uint8:
		r = mimeIds[v]
	default:
		r = nil
	}
	return
}

// Get 获取 string(name/type) uint8(id)
func Get(i any) (h Binder) {
	if t := Type(i); t != nil {
		h = binderMap[t.Type]
	}
	return
}

// Register 注册 Binder。约定仅在启动阶段单线程调用,参见 binderMap 注释。
func Register(t string, handle Binder) error {
	if _, ok := binderMap[t]; ok {
		return fmt.Errorf("handle exist:%v", t)
	}
	binderMap[t] = handle
	return nil
}

func Encode(w io.Writer, i any, t any) error {
	handle := Get(t)
	if handle == nil {
		return errors.New("name not exist")
	}
	return handle.Encode(w, i)
}

func Decode(r io.Reader, i any, t any) error {
	handle := Get(t)
	if handle == nil {
		return errors.New("name not exist")
	}
	return handle.Decode(r, i)
}

func Marshal(i any, t any) ([]byte, error) {
	handle := Get(t)
	if handle == nil {
		return nil, errors.New("type not exist")
	}
	return handle.Marshal(i)
}
func Unmarshal(b []byte, i any, t any) error {
	handle := Get(t)
	if handle == nil {
		return errors.New("type not exist")
	}
	return handle.Unmarshal(b, i)
}
