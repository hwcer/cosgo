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

func (this *Remote) NONE(a int) error {
	return errors.New("remote NONE")
}

func (this *Remote) NONE2(c *cosweb.Context) int {
	return 0
}

func (this *Remote) unExportedFunc(c *cosweb.Context) error {
	return c.String("remote unExportedFunc")
}
