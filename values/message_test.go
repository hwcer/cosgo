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

	// 3. Data 为空
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
