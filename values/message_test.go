package values

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestMessage_Error(t *testing.T) {

	msg := Message{}

	_ = msg.Parse(fmt.Errorf("test error"))
	t.Log(msg.String())

	_ = msg.Parse("test string")
	t.Log(msg.String())

	_ = msg.Parse(100)
	t.Log(msg.String())
	msg.Code = 0

	v := map[string]any{}
	v["k"] = "k"
	v["v"] = 1
	msg.Data = v

	b, err := json.Marshal(msg)
	if err != nil {
		t.Error(err)
	} else {
		t.Logf("test json Marshal:%s", string(b))
	}

	//模拟通过NET获得的message
	r := &Message{}
	if err = json.Unmarshal(b, r); err != nil {
		t.Error(err)
	}

	m := map[string]any{}
	if err = r.Unmarshal(&m); err != nil {
		t.Error(err)
	} else {
		t.Logf("test net Unmarshal:%v", m)
	}

}

// 服务器 → JSON → 客户端 完整链路测试
func TestMessage_Unmarshal(t *testing.T) {
	type Item struct {
		Name  string `json:"name"`
		Count int    `json:"count"`
	}

	// 1. 正常数据：服务器构建 Message，客户端 Unmarshal 到 struct
	t.Run("success_struct", func(t *testing.T) {
		server := &Message{Data: Item{Name: "apple", Count: 3}}
		b, err := json.Marshal(server)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("服务器响应: %s", b)

		client := &Message{}
		if err = json.Unmarshal(b, client); err != nil {
			t.Fatal(err)
		}
		t.Logf("客户端 Data 类型: %T", client.Data)

		var item Item
		if err = client.Unmarshal(&item); err != nil {
			t.Fatal(err)
		}
		if item.Name != "apple" || item.Count != 3 {
			t.Errorf("got %+v", item)
		}
		t.Logf("客户端反序列化: %+v", item)
	})

	// 2. 错误响应：Code!=0，Unmarshal 应返回 error
	t.Run("error_response", func(t *testing.T) {
		server := Errorf(1001, "余额不足")
		b, err := json.Marshal(server)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("服务器错误响应: %s", b)

		client := &Message{}
		if err = json.Unmarshal(b, client); err != nil {
			t.Fatal(err)
		}

		var item Item
		err = client.Unmarshal(&item)
		if err == nil {
			t.Fatal("期望返回错误")
		}
		t.Logf("客户端收到错误: code=%d, err=%v", client.Code, err)
	})

	// 3. 网络反序列化后 Code>0 使用 fmt 打印错误信息（非二进制）
	t.Run("error_string_not_binary", func(t *testing.T) {
		server := Errorf(1001, "余额不足")
		b, err := json.Marshal(server)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("服务器JSON: %s", b)

		client := &Message{}
		if err = json.Unmarshal(b, client); err != nil {
			t.Fatal(err)
		}
		// Data 此时是 json.RawMessage
		t.Logf("Data类型: %T", client.Data)

		got := client.String()
		if got != "余额不足" {
			t.Errorf("String() = %q, want %q", got, "余额不足")
		}
		t.Logf("fmt.Println: %s", client)
		t.Logf("Error(): %s", client.Error())
	})

	// 4. Data 为空
	t.Run("empty_data", func(t *testing.T) {
		server := &Message{Code: 0}
		b, err := json.Marshal(server)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("服务器空响应: %s", b)

		client := &Message{}
		if err = json.Unmarshal(b, client); err != nil {
			t.Fatal(err)
		}

		var item Item
		if err = client.Unmarshal(&item); err != nil {
			t.Fatal(err)
		}
		if item.Name != "" || item.Count != 0 {
			t.Errorf("期望零值, got %+v", item)
		}
		t.Logf("客户端空数据: %+v", item)
	})
}

// TestMessage_UnmarshalKeepsCode 🔴 Code 必须活着穿过 Unmarshal。
//
// 这是 RPC 的返回路径（cosrpc/client.XCall 在 reply 非 nil 时就调 Unmarshal），
// 所以本函数丢掉 Code 等于**全链路的业务错误码都被抹平成默认码**。
//
// 2026-09-02 真出过：公会那一整批业务码（7001~7024）一条都没到客户端，症状是
// 「文案对、码变 9999」，与「压根没配码」一模一样，极难往这一跳上想。
//
// ⚠ 同文件的 error_response 用例只断言 `err != nil` —— 那条**改前改后都过**，
// 正是这种空洞守卫让缺陷活了下来。判据必须落在 Code 上。
func TestMessage_UnmarshalKeepsCode(t *testing.T) {
	type Item struct {
		Name string `json:"name"`
	}
	b, err := json.Marshal(Errorf(7011, "职位权限不足"))
	if err != nil {
		t.Fatal(err)
	}
	client := &Message{}
	if err = json.Unmarshal(b, client); err != nil {
		t.Fatal(err)
	}

	var item Item
	err = client.Unmarshal(&item)
	if err == nil {
		t.Fatal("Code 非 0 时必须返回错误")
	}
	//文案与旧写法（errors.New(文案)）逐字一致 —— 修这条不该改变任何人看到的文字
	if err.Error() != "职位权限不足" {
		t.Fatalf("Error() = %q, want %q", err.Error(), "职位权限不足")
	}
	//🔴 关键断言：错误里必须还带着码
	m, ok := err.(*Message)
	if !ok {
		t.Fatalf("返回的错误应是 *Message（带 Code），实得 %T —— 码在这里丢了", err)
	}
	if m.Code != 7011 {
		t.Fatalf("Code = %d, want 7011", m.Code)
	}
}

// TestMessage_UnmarshalNonStringData Data 不是 JSON 字符串时也要保住码与文案。
//
// 旧写法先 json.Unmarshal 进 string、失败就把**解码错误**当业务错误返回 —— 码和文案一起丢，
// 客户端连出了什么事都认不出来。任何把数字、对象等非字符串放进 Data 的错误回包都会踩到。
func TestMessage_UnmarshalNonStringData(t *testing.T) {
	b, err := json.Marshal(&Message{Code: 209, Data: 30})
	if err != nil {
		t.Fatal(err)
	}
	client := &Message{}
	if err = json.Unmarshal(b, client); err != nil {
		t.Fatal(err)
	}
	var out any
	err = client.Unmarshal(&out)
	m, ok := err.(*Message)
	if !ok {
		t.Fatalf("应回 *Message，实得 %T", err)
	}
	if m.Code != 209 {
		t.Fatalf("Code = %d, want 209", m.Code)
	}
	if m.Error() != "30" {
		t.Fatalf("Error() = %q, want %q（剩余秒数应原样可读）", m.Error(), "30")
	}
}

// TestMessage_ArgsSurviveJSON 🔴 Args 必须活着穿过一次完整的序列化往返。
//
// 守的是 rawMessage 漏字段那个坑：UnmarshalJSON 不走默认反射，而是先解进 rawMessage
// 再逐字段搬运，任何忘了搬的字段都在**每一次 RPC 回程**被静默丢掉，而且不报错、
// 症状只是「客户端拿不到参数」，与「服务器压根没填」完全一样，极难往这一跳上想。
//
// ⚠ 判据必须落在 Args 的**值**上。断言 `Args != nil` 是空洞守卫 —— 真出问题时它是 nil，
// 但更常见的退化是长度对、内容错（比如嵌套成 [[1001 5 2]]），那种断言照样放过去。
func TestMessage_ArgsSurviveJSON(t *testing.T) {
	server := Errorf(1002, "Item Not Enough:%v", 1001).WithArgs(1001, 5, 2)

	b, err := json.Marshal(server)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("服务器错误响应: %s", b)

	client := &Message{}
	if err = json.Unmarshal(b, client); err != nil {
		t.Fatal(err)
	}
	if client.Code != 1002 {
		t.Fatalf("Code = %d, want 1002", client.Code)
	}
	if len(client.Args) != 3 {
		t.Fatalf("Args = %v, want 长度 3 —— 字段大概率没搬进 rawMessage", client.Args)
	}
	//过一次 JSON 之后数字是 float64，Go 侧按约定用 ParseInt32 取值
	want := []int32{1001, 5, 2}
	for i, w := range want {
		if got := ParseInt32(client.Args[i]); got != w {
			t.Fatalf("Args[%d] = %v(%T), want %d", i, client.Args[i], client.Args[i], w)
		}
	}
	t.Logf("客户端拿到道具参数: iid=%v need=%v have=%v", client.Args[0], client.Args[1], client.Args[2])
}

// TestMessage_ArgsOmittedWhenEmpty 无参消息的线上报文必须一字节不变。
//
// 心跳、重连回包这类 Message{Code:0, Data:…} 在长连接上每秒都在发，
// 加字段不能让这些热路径的包变大，也不能让老客户端见到没见过的键。
func TestMessage_ArgsOmittedWhenEmpty(t *testing.T) {
	b, err := json.Marshal(&Message{Code: 0, Data: 1700000000000})
	if err != nil {
		t.Fatal(err)
	}
	if s := string(b); s != `{"code":0,"data":1700000000000}` {
		t.Fatalf("心跳报文 = %s，多出了字段", s)
	}
}

// TestErrorf_SentinelNotMutated 🔴 Errorf 不得改写传进来的包级哨兵。
//
// 上层惯用 var ErrXxx = Errorf(...) 在 init 期建一份哨兵、全进程复用，那是**进程内唯一的
// 共享指针**。Errorf 从前直接在参数上改 Code，加了 Args 之后还要改 Args——
// 那是跨 goroutine 的全局写：一次 Errorf(500, ErrXxx) 就能把哨兵的码永久改掉，
// 之后所有请求拿到的都是被污染的值，而且污染只在特定调用顺序下出现，测不出、看不见。
func TestErrorf_SentinelNotMutated(t *testing.T) {
	sentinel := Errorf(404, "page not found")

	got := Errorf(0, sentinel, 1001)

	if sentinel.Args != nil {
		t.Fatalf("哨兵被写入了 Args = %v —— 全局状态被污染", sentinel.Args)
	}
	if got == sentinel {
		t.Fatal("需要改写字段时必须返回副本，不能返回哨兵本身")
	}
	if len(got.Args) != 1 || ParseInt32(got.Args[0]) != 1001 {
		t.Fatalf("副本 Args = %v, want [1001]", got.Args)
	}
	if got.Code != 404 {
		t.Fatalf("Code = %d, want 404（未指定新码时沿用哨兵的码）", got.Code)
	}
	if got.Error() != "page not found" {
		t.Fatalf("Error() = %q, want %q（文案不该被改动）", got.Error(), "page not found")
	}

	//不需要改写任何字段时保持原指针，不平白多一次分配
	if same := Errorf(0, sentinel); same != sentinel {
		t.Fatal("无改写时应原样返回哨兵指针")
	}
}
