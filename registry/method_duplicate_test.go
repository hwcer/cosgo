package registry

import (
	"testing"
)

// TestMethodDuplicate 测试方法重复注册
func TestMethodDuplicate(t *testing.T) {
	// 创建路由器
	router := NewRouter()

	// 创建测试节点
	staticNode := &Node{name: "/test/static"}
	paramNode := &Node{name: "/test/:id"}

	// 注册路由
	err := router.Register(staticNode, []string{"GET"})
	if err != nil {
		t.Fatalf("注册静态路由失败: %v", err)
	}

	err = router.Register(paramNode, []string{"GET"})
	if err != nil {
		t.Fatalf("注册参数路由失败: %v", err)
	}

	// 测试重复注册同一方法到静态路由
	t.Run("重复注册静态路由方法", func(t *testing.T) {
		err := router.Register(staticNode, []string{"GET"})
		if err == nil {
			t.Fatal("重复注册静态路由方法应该返回错误")
		}
		t.Logf("重复注册静态路由方法返回错误: %v", err)
	})

	// 测试重复注册同一方法到参数路由
	t.Run("重复注册参数路由方法", func(t *testing.T) {
		err := router.Register(paramNode, []string{"GET"})
		if err == nil {
			t.Fatal("重复注册参数路由方法应该返回错误")
		}
		t.Logf("重复注册参数路由方法返回错误: %v", err)
	})

	// 测试注册不同方法到同一路由
	t.Run("注册不同方法到同一路由", func(t *testing.T) {
		err := router.Register(staticNode, []string{"POST"})
		if err != nil {
			t.Fatalf("注册不同方法到同一路由失败: %v", err)
		}
		t.Logf("注册不同方法到同一路由成功")
	})
}
