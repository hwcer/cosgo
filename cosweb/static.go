package cosweb

import (
	"cosgo/app"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

const iStaticRoutePath = "_StaticRoutePath"

type Static struct {
	root string
}

func NewStatic(root string) *Static {
	if !path.IsAbs(root) {
		root = filepath.Join(app.Config.GetString("appWorkDir"), root)
	}
	return &Static{root: root}
}

func (this *Static) Route(s *Server, prefix string, method ...string) {
	arr := []string{strings.TrimSuffix(prefix, "/"), "*" + iStaticRoutePath}
	route := strings.Join(arr, "/")
	s.Register(route, this.handle, method...)
}

func (this *Static) handle(c *Context) error {
	name := c.Get(iStaticRoutePath, RequestDataTypePath)
	if name == "" {
		return nil
	}
	file := filepath.Join(this.root, name)
	if !strings.HasPrefix(file, this.root) {
		return nil
	}
	http.ServeFile(c.Response, c.Request, file)
	return nil
}
