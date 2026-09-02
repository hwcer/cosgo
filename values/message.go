package values

import (
	"encoding/json"
	"fmt"
)

type rawMessage struct {
	Code int32            `json:"code"`
	Data json.RawMessage  `json:"data"`
}

const MessageErrorCodeDefault int32 = 9999

type Message struct {
	Code int32 `json:"code"`
	Data any   `json:"data"`
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
// 顺带修掉旧写法的一处窄坑:Data 不是 JSON 字符串时(顶号回包的 Data 是剩余秒数),
// 旧写法的 json.Unmarshal(v, &s) 会失败,于是把**解码错误**当业务错误返回、连文案都丢;
// 现在原样返回 Message,String() 自己回落到 string(v)。
func (this *Message) Unmarshal(i interface{}) (err error) {
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
		if code != 0 {
			r.Code = code
		} else if r.Code == 0 {
			r.Code = MessageErrorCodeDefault
		}
		return r
	}
	r = &Message{}
	r.Errorf(code, format, args...)
	return r
}
