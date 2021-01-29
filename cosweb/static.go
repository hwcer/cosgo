package cosweb

import (
	"cosgo/apps"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

const iStaticRoutePath = "*"

type Static struct {
	root string
}

func NewStatic(root string) *Static {
	if !path.IsAbs(root) {
		root = filepath.Join(apps.Config.GetString("appWorkDir"), root)
	}
	return &Static{root: root}
}

func (this *Static) handler(c *Context) error {
	if len(c.values) < 1 {
		return nil
	}
	name := c.values[len(c.values)-1]
	file := filepath.Join(this.root, name)
	if !strings.HasPrefix(file, this.root) {
		return nil
	}
	http.ServeFile(c.Response, c.Request, file)
	return nil
}
