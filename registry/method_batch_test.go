package registry

import (
	"testing"
)

// TestMethodBatch 测试批量注册方法
func TestMethodBatch(t *testing.T) {
	// 创建路由器
	router := NewRouter()

	// 创建测试节点
	paramNode := &Node{name: "/test/:id"}

	// 测试批量注册多个方法
	t.Run("批量注册多个方法", func(t *testing.T) {
		err := router.Register(paramNode, []string{"GET", "POST", "PUT"})
		if err != nil {
			t.Fatalf("批量注册多个方法失败: %v", err)
		}
		t.Log("批量注册多个方法成功")

		// 验证所有方法都已注册
		node, _ := router.Search("GET", "test", "123")
		if node == nil {
			t.Fatal("GET方法未注册")
		}

		node, _ = router.Search("POST", "test", "123")
		if node == nil {
			t.Fatal("POST方法未注册")
		}

		node, _ = router.Search("PUT", "test", "123")
		if node == nil {
			t.Fatal("PUT方法未注册")
		}
	})

	// 测试批量注册包含重复方法
	t.Run("批量注册包含重复方法", func(t *testing.T) {
		newNode := &Node{name: "/test/:id/new"}
		err := router.Register(newNode, []string{"GET", "GET"})
		if err == nil {
			t.Fatal("批量注册包含重复方法应该返回错误")
		}
		t.Logf("批量注册包含重复方法返回错误: %v", err)
	})

	// 测试批量注册包含已存在方法
	t.Run("批量注册包含已存在方法", func(t *testing.T) {
		newNode := &Node{name: "/test/:id"}
		err := router.Register(newNode, []string{"GET", "DELETE"})
		if err == nil {
			t.Fatal("批量注册包含已存在方法应该返回错误")
		}
		t.Logf("批量注册包含已存在方法返回错误: %v", err)
	})
}
