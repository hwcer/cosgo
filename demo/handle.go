package main

import (
	"cosgo/express"
	"errors"
)

type remote struct {
}

func (this *remote) Test(c *express.Context) error {
	return c.String("remote test")
}

func (this *remote) Error(c *express.Context) error {
	return errors.New("remote error")
}

func (this *remote) NONE(a int) error {
	return errors.New("remote NONE")
}

func (this *remote) NONE2(c *express.Context) int {
	return 0
}

func (this *remote) unExportedFunc(c *express.Context) error {
	return c.String("remote unExportedFunc")
}
