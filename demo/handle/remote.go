package handle

import (
	"cosgo/express"
	"errors"
)

type Remote struct {
}

func (this *Remote) Test(c *express.Context) error {
	return c.String("remote test")
}

func (this *Remote) Error(c *express.Context) error {
	return errors.New("remote error")
}

func (this *Remote) NONE(a int) error {
	return errors.New("remote NONE")
}

func (this *Remote) NONE2(c *express.Context) int {
	return 0
}

func (this *Remote) unExportedFunc(c *express.Context) error {
	return c.String("remote unExportedFunc")
}
