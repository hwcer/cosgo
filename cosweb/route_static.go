package cosweb

import (
	"github.com/hwcer/cosgo/app"
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

func (this *Static) handle(c *Context, next Next) (err error) {
	name := c.Get(iStaticRoutePath, RequestDataTypeParam)
	file := filepath.Join(this.root, name)
	//fmt.Printf("static file:%v\n", file)
	http.ServeFile(c.Response, c.Request, file)
	return
}
