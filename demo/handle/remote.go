package handle

import (
	"errors"
	"github.com/hwcer/cosgo/cosweb"
)

type Remote struct {
	v int
}

func (this *Remote) Tpl(c *cosweb.Context) error {
	v := make(map[string]interface{})
	v["name"] = "模板测试"
	return c.Render("index.html", v)
}

func (this *Remote) Error(c *cosweb.Context) error {
	return errors.New("remote error")
}

func (this *Remote) Addr(c *cosweb.Context) error {
	return c.String(c.RemoteAddr())
}
func (this *Remote) Jsonp(c *cosweb.Context) error {
	k := c.Get("callback")
	d := make(map[string]string)
	d["a"] = "x"
	d["b"] = "y"
	return c.JSONP(k, d)
}
