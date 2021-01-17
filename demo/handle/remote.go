package handle

import (
	"cosgo/cosweb"
	"errors"
)

type Remote struct {
}

func (this *Remote) Test(c *cosweb.Context) error {
	return c.String("remote test")
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
