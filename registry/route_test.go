package registry

import (
	"testing"
)

// TestUIRouteMatching 测试 /ui/* 路由匹配
func TestUIRouteMatching(t *testing.T) {
	// 创建路由器
	r := NewRouter()

	// 注册路由: /*
	node1 := &Node{name: "/*"}
	err := r.Register(node1, []string{"GET"})
	if err != nil {
		t.Fatalf("注册路由1失败: %v", err)
	}

	// 注册路由: /ui/*
	node2 := &Node{name: "/ui/*"}
	err = r.Register(node2, []string{"GET"})
	if err != nil {
		t.Fatalf("注册路由2失败: %v", err)
	}

	// 测试用例
	testCases := []struct {
		path     string
		expected string
		name     string
	}{
		{
			path:     "/",
			expected: "/*",
			name:     "访问 / 应该匹配 /*",
		},
		{
			path:     "/ui",
			expected: "/ui/*",
			name:     "访问 /ui 应该匹配 /ui/*",
		},
		{
			path:     "/ui/index",
			expected: "/ui/*",
			name:     "访问 /ui/index 应该匹配 /ui/*",
		},
		{
			path:     "/test",
			expected: "/*",
			name:     "访问 /test 应该匹配 /*",
		},
	}

	// 执行测试
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n, _ := r.Search("GET", tc.path)
			if n == nil {
				t.Errorf("路径 %s 没有匹配到任何路由", tc.path)
				return
			}
			if n.Name() != tc.expected {
				t.Errorf("路径 %s 应该匹配 %s，但实际匹配到了 %s", tc.path, tc.expected, n.Name())
			} else {
				t.Logf("路径 %s 匹配成功，匹配到路由: %s", tc.path, n.Name())
			}
		})
	}
}
