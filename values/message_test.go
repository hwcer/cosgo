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

	v := map[string]interface{}{}
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

	m := map[string]interface{}{}
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
// 顶号回包的 Data 是**剩余秒数**（数字）。旧写法先 json.Unmarshal 进 string、失败就
// 把解码错误当业务错误返回 —— 码和文案一起丢，客户端连「被顶号了」都认不出来。
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
