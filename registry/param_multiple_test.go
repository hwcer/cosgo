package registry

import (
	"testing"
)

// TestParamMultiple 测试多个参数名
func TestParamMultiple(t *testing.T) {
	// 创建路由器
	router := NewRouter()

	// 创建测试节点
	paramNode1 := &Node{name: "/test/:id"}
	paramNode2 := &Node{name: "/test/:user"}

	// 注册路由
	err := router.Register(paramNode1, []string{"GET"})
	if err != nil {
		t.Fatalf("注册参数路由1失败: %v", err)
	}

	err = router.Register(paramNode2, []string{"POST"})
	if err != nil {
		t.Fatalf("注册参数路由2失败: %v", err)
	}

	// 测试参数路由（多个参数名）
	t.Run("多个参数名", func(t *testing.T) {
		// 测试 GET 方法
		node, params := router.Search("GET", "test", "USER123")
		if node == nil {
			t.Fatal("GET 参数路由未找到")
		}
		if len(params) != 2 {
			t.Fatalf("GET 参数路由应返回2个参数: %v", params)
		}
		if params["id"] != "USER123" {
			t.Fatalf("GET 参数id值错误: 期望 USER123, 实际 %s", params["id"])
		}
		if params["user"] != "USER123" {
			t.Fatalf("GET 参数user值错误: 期望 USER123, 实际 %s", params["user"])
		}

		// 测试 POST 方法
		node, params = router.Search("POST", "test", "USER123")
		if node == nil {
			t.Fatal("POST 参数路由未找到")
		}
		if len(params) != 2 {
			t.Fatalf("POST 参数路由应返回2个参数: %v", params)
		}
		if params["id"] != "USER123" {
			t.Fatalf("POST 参数id值错误: 期望 USER123, 实际 %s", params["id"])
		}
		if params["user"] != "USER123" {
			t.Fatalf("POST 参数user值错误: 期望 USER123, 实际 %s", params["user"])
		}
	})
}
