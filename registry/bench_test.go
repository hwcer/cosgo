package registry

import (
	"testing"
)

type benchHandler struct{}

func (benchHandler) Filter(*Node) bool { return true }

func buildBenchRouter() *Registry {
	r := New()
	svc := r.Service("api", benchHandler{})
	svc.SetMethods([]string{"GET"})
	// 注册一些静态路由
	for _, p := range []string{"/api/users", "/api/health", "/api/status"} {
		node := &Node{name: p, service: svc}
		_ = svc.router.Register(node, []string{"GET"})
	}
	// 一条参数路由 /api/users/:id
	nodeParam := &Node{name: "/api/users/:id", service: svc}
	_ = svc.router.Register(nodeParam, []string{"GET"})
	// 一条通配符路由 /assets/*path
	nodeWild := &Node{name: "/assets/*path", service: svc}
	_ = svc.router.Register(nodeWild, []string{"GET"})
	return r
}

func BenchmarkSearch_StaticHit(b *testing.B) {
	r := buildBenchRouter()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Search("GET", "/api/users")
	}
}

func BenchmarkSearch_ParamHit(b *testing.B) {
	r := buildBenchRouter()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Search("GET", "/api/users/42")
	}
}

func BenchmarkSearch_WildcardHit(b *testing.B) {
	r := buildBenchRouter()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Search("GET", "/assets/css/app.min.css")
	}
}

func BenchmarkSearch_NotFound(b *testing.B) {
	r := buildBenchRouter()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = r.Search("GET", "/nope/not/here")
	}
}

func BenchmarkJoin_FastPath(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Join("/api/users/42")
	}
}
