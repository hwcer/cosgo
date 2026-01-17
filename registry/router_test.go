package registry

import (
	"testing"
)

// TestRouter 测试路由功能
func TestRouter(t *testing.T) {
	// 创建路由器
	router := NewRouter()

	// 创建测试节点
	staticNode := &Node{name: "/test/static"}
	paramNode := &Node{name: "/test/:id"}
	paramSubmitNode := &Node{name: "/test/:id/submit"}
	wildNode := &Node{name: "/test/*"}
	rootNode := &Node{name: "/"}
	globalWildNode := &Node{name: "/*"}

	// 注册路由
	err := router.Register(staticNode, []string{"GET"})
	if err != nil {
		t.Fatalf("注册静态路由失败: %v", err)
	}

	err = router.Register(paramNode, []string{"GET"})
	if err != nil {
		t.Fatalf("注册参数路由失败: %v", err)
	}

	err = router.Register(wildNode, []string{"GET"})
	if err != nil {
		t.Fatalf("注册通配符路由失败: %v", err)
	}

	err = router.Register(paramSubmitNode, []string{"GET"})
	if err != nil {
		t.Fatalf("注册参数提交路由失败: %v", err)
	}

	err = router.Register(rootNode, []string{"GET"})
	if err != nil {
		t.Fatalf("注册根路径路由失败: %v", err)
	}

	err = router.Register(globalWildNode, []string{"GET"})
	if err != nil {
		t.Fatalf("注册全局通配符路由失败: %v", err)
	}

	// 测试静态路由
	t.Run("静态路由", func(t *testing.T) {
		node, params := router.Search("GET", "test", "static")
		if node == nil {
			t.Fatal("静态路由未找到")
		}
		if node.Name() != "/test/static" {
			t.Fatalf("静态路由节点名称错误: 期望 /test/static, 实际 %s", node.Name())
		}
		if params != nil && len(params) > 0 {
			t.Fatalf("静态路由不应返回参数: %v", params)
		}
	})

	// 测试参数路由
	t.Run("参数路由", func(t *testing.T) {
		node, params := router.Search("GET", "test", "123")
		if node == nil {
			t.Fatal("参数路由未找到")
		}
		if node.Name() != "/test/:id" {
			t.Fatalf("参数路由节点名称错误: 期望 /test/:id, 实际 %s", node.Name())
		}
		if params == nil || len(params) != 1 {
			t.Fatalf("参数路由应返回参数: %v", params)
		}
		if params["id"] != "123" {
			t.Fatalf("参数值错误: 期望 123, 实际 %s", params["id"])
		}
	})

	// 测试参数提交路由
	t.Run("参数提交路由", func(t *testing.T) {
		node, params := router.Search("GET", "test", "456", "submit")
		if node == nil {
			t.Fatal("参数提交路由未找到")
		}
		if node.Name() != "/test/:id/submit" {
			t.Fatalf("参数提交路由节点名称错误: 期望 /test/:id/submit, 实际 %s", node.Name())
		}
		if params == nil || len(params) != 1 {
			t.Fatalf("参数提交路由应返回参数: %v", params)
		}
		if params["id"] != "456" {
			t.Fatalf("参数值错误: 期望 456, 实际 %s", params["id"])
		}
	})

	// 测试通配符路由
	t.Run("通配符路由", func(t *testing.T) {
		node, _ := router.Search("GET", "test", "path", "to", "file")
		if node == nil {
			t.Fatal("通配符路由未找到")
		}
		if node.Name() != "/test/*" {
			t.Fatalf("通配符路由节点名称错误: 期望 /test/*, 实际 %s", node.Name())
		}
	})

	// 测试路由优先级
	t.Run("路由优先级", func(t *testing.T) {
		// 测试更具体的路由优先于全局通配符路由
		node, _ := router.Search("GET", "test", "static")
		if node == nil {
			t.Fatal("静态路由未找到")
		}
		if node.Name() != "/test/static" {
			t.Fatalf("静态路由应优先于全局通配符路由: 期望 /test/static, 实际 %s", node.Name())
		}
	})

	// 测试不同HTTP方法
	t.Run("不同HTTP方法", func(t *testing.T) {
		// 注册POST方法的静态路由
		postNode := &Node{name: "/api/post"}
		err := router.Register(postNode, []string{"POST"})
		if err != nil {
			t.Fatalf("注册POST方法路由失败: %v", err)
		}

		// 测试POST方法访问POST路由
		node, _ := router.Search("POST", "api", "post")
		if node == nil {
			t.Fatal("POST方法应访问POST路由")
		}
		if node.Name() != "/api/post" {
			t.Fatalf("POST路由节点名称错误: 期望 /api/post, 实际 %s", node.Name())
		}
	})

	// 测试根路径路由
	t.Run("根路径路由", func(t *testing.T) {
		node, params := router.Search("GET")
		if node == nil {
			t.Fatal("根路径路由未找到")
		}
		if node.Name() != "/" {
			t.Fatalf("根路径路由节点名称错误: 期望 /, 实际 %s", node.Name())
		}
		if params != nil && len(params) > 0 {
			t.Fatalf("根路径路由不应返回参数: %v", params)
		}
	})

	// 测试全局通配符路由
	t.Run("全局通配符路由", func(t *testing.T) {
		// 测试未匹配到其他路由的路径
		node, _ := router.Search("GET", "any", "path", "here")
		if node == nil {
			t.Fatal("全局通配符路由未找到")
		}
		if node.Name() != "/*" {
			t.Fatalf("全局通配符路由节点名称错误: 期望 /*, 实际 %s", node.Name())
		}
	})
}
