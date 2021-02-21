package handle

import (
	"cosgo/cosweb"
	"errors"
	"strconv"
)

type Remote struct {
	v int
}

func (this *Remote) Test(c *cosweb.Context) error {
	this.v++
	return c.String(strconv.Itoa(this.v))
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
