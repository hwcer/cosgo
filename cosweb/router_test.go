package cosweb

import (
	"testing"
)

var router *Router

func init() {
	router = NewRouter()

	router.Register([]string{"POST"}, "/", nil)
	router.Register([]string{"POST"}, "/*", nil)
	router.Register([]string{"POST"}, "/a/*", nil)
	router.Register([]string{"POST"}, "/a/b/*", nil)
	router.Register([]string{"POST"}, "/:a/b/c", nil)
	router.Register([]string{"POST"}, "/a/:b/c", nil)
	router.Register([]string{"POST"}, "/:a/:b/c", nil)
	router.Register([]string{"POST"}, "/:a/:b/:c", nil)
	router.Register([]string{"POST"}, "/a/b/c", nil)
}

func TestRoute(t *testing.T) {
	matchT("/", t)
	matchT("/a", t)
	matchT("/a/b/c", t)
	matchT("/x/b/c", t)
	matchT("/x/y/c", t)
	matchT("/a/y/c", t)
	matchT("/x/y/z", t)
}
func BenchmarkRoute(b *testing.B) {
	matchB("/", b)
	matchB("/a", b)
	matchB("/a/b/c", b)
	matchB("/x/b/c", b)
	matchB("/x/y/c", b)
	matchB("/a/y/c", b)
	matchB("/x/y/z", b)
}
func matchT(path string, t *testing.T) {
	node := router.Match("post", path)
	if node != nil {
		t.Logf("匹配成功：%v --> %v , args:%+v", path, node.String(), node.Params(path))
	} else {
		t.Logf("匹配失败：%v ", path)
	}
}

func matchB(path string, t *testing.B) {
	node := router.Match("post", path)
	if node != nil {
		t.Logf("匹配成功：%v --> %v , args:%+v", path, node.String(), node.Params(path))
	} else {
		t.Logf("匹配失败：%v ", path)
	}
}
