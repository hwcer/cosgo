package cosweb

import (
	"net/http"
	"testing"
)

var router *Router

func test(c *Context, f func()) error {
	return nil
}
func init() {
	router = NewRouter()

	router.Register("/", test, http.MethodGet)
	router.Register("/*", test, http.MethodGet)
	router.Register("/a/*", test, http.MethodGet)
	router.Register("/a/b/*", test, http.MethodGet)
	router.Register("/:a/b/c", test, http.MethodGet)
	router.Register("/a/:b/c", test, http.MethodGet)
	router.Register("/:a/:b/c", test, http.MethodGet)
	router.Register("/:a/:b/:c", test, http.MethodGet)
	router.Register("/a/b/c", test, http.MethodGet)
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
	node := router.Match(http.MethodGet, path)
	if len(node) > 0 {
		t.Logf("匹配成功：%v  , nodes:%+v", path, node)
	} else {
		t.Logf("匹配失败：%v ", path)
	}
}

func matchB(path string, t *testing.B) {
	node := router.Match(http.MethodGet, path)
	if node != nil {
		t.Logf("匹配成功：%v  , nodes:%+v", path, node)
	} else {
		t.Logf("匹配失败：%v ", path)
	}
}
