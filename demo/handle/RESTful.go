package handle

import (
	"cosgo/express"
)

type Restful struct {
}

func (this *Restful) GET(c *express.Context) error {
	return c.String("Restful Get")
}

func (this *Restful) PUT(c *express.Context) error {
	return c.String("Restful PUT")
}

func (this *Restful) POST(c *express.Context) error {
	return c.String("Restful POST")
}
func (this *Restful) DELETE(c *express.Context) error {
	return c.String("Restful DELETE")
}
