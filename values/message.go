package values

import (
	"encoding/json"
	"fmt"
)

// rawMessage UnmarshalJSON 的影子结构。
// 🔴 Message 每加一个字段,这里必须同步加,否则该字段在每一次反序列化(即每一次 RPC 回程、
// 每一次客户端解包)被静默丢掉 —— 与 Unmarshal 上方注释记录的 2026-09-02 丢码事故同一类。
type rawMessage struct {
	Code int32           `json:"code"`
	Data json.RawMessage `json:"data"`
	Args []any           `json:"args,omitempty"`
}

const MessageErrorCodeDefault int32 = 9999

type Message struct {
	Code int32 `json:"code"`
	Data any   `json:"data"`
	// Args 错误参数,客户端据此定位出错的具体对象(道具ID、需要数量等),不必去解 Data 里的文案。
	// 位置参数:同一个 Code 每次给出的参数个数与含义必须一致,由定义该 Code 的那一层写进文档,
	// 本层不做任何约束。无参时 omitempty 不落到报文里,心跳这类无参消息的线上字节数不变。
	//
	// ⚠ 过一次 JSON 之后数字一律变 float64(1001 → float64(1001))。
	// Go 侧消费方用本包的 ParseInt32/ParseInt64 转换,不要直接做 .(int32) 断言。
	Args []any `json:"args,omitempty"`
}

func (this *Message) Parse(v any) *Message {
	switch v.(type) {
	case error:
		this.Errorf(0, v)
	default:
		this.Data = v
	}
	return this
}
func (this *Message) String() string {
	if this.Data == nil {
		return ""
	}
	switch v := this.Data.(type) {
	case string:
		return v
	case json.RawMessage:
		var s string
		if json.Unmarshal(v, &s) == nil {
			return s
		}
		return string(v)
	case []byte:
		return string(v)
	default:
		return fmt.Sprintf("%v", this.Data)
	}
}

func (this *Message) Error() string {
	return this.String()
}

// Errorf 格式化一个错误,必定产生错误码
func (this *Message) Errorf(code int32, format any, args ...any) {
	if code == 0 {
		this.Code = MessageErrorCodeDefault
	} else {
		this.Code = code
	}
	this.Data = Sprintf(format, args...)
	//无条件赋值:Errorf 是在重新定义整个错误,Args 与 Data 同进同出,
	//避免复用同一个 Message 时残留上一次的参数。
	this.Args = args
}

// WithArgs 覆盖错误参数,返回自身便于链式调用。
//
// 用于「格式化实参 ≠ 语义参数」的场合:Errorf 会把收到的 args 原样灌进 Args,
// 但 func ErrXxx(args ...any) 这类变参包装往往把整个切片当**一个** %v 实参传下来,
// 自动灌入得到的是嵌套的 [[1001 5 2]],需要在这里拍平成 [1001 5 2]。
func (this *Message) WithArgs(args ...any) *Message {
	this.Args = args
	return this
}

func (this *Message) UnmarshalJSON(b []byte) error {
	if _, ok := this.Data.(json.RawMessage); ok {
		return nil
	}
	var raw rawMessage
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}
	this.Code = raw.Code
	this.Data = raw.Data
	this.Args = raw.Args
	return nil
}

// Unmarshal 反序列化Data，如果Code!=0则返回错误信息
// Unmarshal 把回包解进调用方要的结构;Code 非 0 时把**本 Message 自己**当错误返回。
//
// # 🔴 错误分支必须 return this,不能 errors.New(文案)
//
// 后者是本函数从前的写法,它**把错误码扔了**。而这里是 RPC 的返回路径
// (cosrpc/client.XCall 在 reply 非 nil 时就调本函数),于是**任何跨进程返回的业务
// 错误码都在这一跳被抹平**,调用方只剩一个裸 error:
//
//	social 返回 Message{Code:7011, Data:"职位权限不足"}
//	  → errors.New("职位权限不足")   码没了
//	  → 上层按「不带码的普通 error」处置 → 落到默认码 9999
//	  → 玩家界面:ErrDefault(9999): 职位权限不足
//
// 2026-09-02 实测:公会新加的那一整批业务码(7001~7024)一条都没到客户端,
// 症状与「压根没配码」完全一样 —— **文案对、只有码错**,极难往这一跳上想。
// 对照组是同一个接口里游戏服本地返回的 ErrArgEmpty:那条不过 RPC,码 123 好好的。
//
// `*Message` 本身满足 error,且 Error() 走 String(),对 json.RawMessage 会解出那个
// 字符串 —— 所以**错误文案与旧写法逐字一致**,只是多带了 Code,调用方不受影响。
//
// 顺带修掉旧写法的一处窄坑:Data 不是 JSON 字符串时(数字、对象等),
// 旧写法的 json.Unmarshal(v, &s) 会失败,于是把**解码错误**当业务错误返回、连文案都丢;
// 现在原样返回 Message,String() 自己回落到 string(v)。
func (this *Message) Unmarshal(i any) (err error) {
	switch v := this.Data.(type) {
	case json.RawMessage:
		if this.Code != 0 {
			return this
		}
		if len(v) > 0 {
			err = json.Unmarshal(v, i)
		}
	case []byte:
		if this.Code != 0 {
			return this
		}
		if len(v) > 0 {
			err = json.Unmarshal(v, i)
		}
	}
	return
}

func Parse(v any) *Message {
	switch d := v.(type) {
	case *Message:
		return d
	case Message:
		return &d
	default:
		r := &Message{}
		return r.Parse(v)
	}
}

func Error(err any) (r *Message) {
	return Errorf(0, err)
}
func Errorf(code int32, format any, args ...any) (r *Message) {
	switch v := format.(type) {
	case *Message:
		r = v
	case Message:
		r = &v
	}
	if r != nil {
		//🔴 写时复制。传进来的 *Message 极可能是包级共享哨兵 —— 上层惯用
		//var ErrXxx = Errorf(...) 在 init 期建一份、全进程复用。直接往上面写 Code 或 Args
		//就是跨 goroutine 改全局:一次 Errorf(500, ErrXxx) 能把那个哨兵的码永久改掉,
		//之后所有人拿到的都是被污染的值。
		//不需要写任何字段时保持返回原指针,不平白多一次分配。
		if code != 0 || r.Code == 0 || len(args) > 0 {
			v := *r
			r = &v
		}
		if code != 0 {
			r.Code = code
		} else if r.Code == 0 {
			r.Code = MessageErrorCodeDefault
		}
		if len(args) > 0 {
			r.Args = args
		}
		return r
	}
	r = &Message{}
	r.Errorf(code, format, args...)
	return r
}
