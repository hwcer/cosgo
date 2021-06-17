package handle

import (
	"github.com/hwcer/cosgo/cosweb"
)

type Restful struct {
}

func (this *Restful) GET(c *cosweb.Context) error {
	return c.String("Restful Get")
}

func (this *Restful) PUT(c *cosweb.Context) error {
	return c.String("Restful PUT")
}

func (this *Restful) POST(c *cosweb.Context) error {
	return c.String("Restful POST")
}
func (this *Restful) DELETE(c *cosweb.Context) error {
	return c.String("Restful DELETE")
}
