package cosweb

import (
	"testing"
)

var router *Router

func init() {
	router = NewRouter()

	router.Register([]string{"POST"}, "/", nil)
	router.Register([]string{"POST"}, "/c/*argx", nil)
	router.Register([]string{"POST"}, "/:arg1/x", nil)
	router.Register([]string{"POST"}, "/:arg2/y", nil)
	router.Register([]string{"POST"}, "/a/x", nil)
}

func TestRoute(t *testing.T) {
	match("/", t)
	match("/a", t)
	match("/a/b/c", t)
	match("/a/x", t)
	match("/a/y", t)
	match("/b/x", t)
	match("/b/y", t)
	match("/c/x/y", t)
}

func match(path string, t *testing.T) {
	node := router.Match("post", path)
	if node != nil {
		t.Logf("匹配成功：%v --> %v , args:%+v", path, node.String(), node.Params(path))
	} else {
		t.Logf("匹配失败：%v ", path)
	}
}
