package handle

import (
	"errors"
	"github.com/hwcer/cosgo/cosweb"
	"time"
)

type Admin struct {
}

func (this *Admin) Login(c *cosweb.Context) error {
	name := c.Get("name")
	if name == "" {
		return c.String("name empty")
	}
	m := make(map[string]interface{})
	m["name"] = name
	sid, err := c.Session.Create(m)
	if err != nil {
		return err
	}
	return c.String(sid)
}

func (this *Admin) Show(c *cosweb.Context) error {
	return c.JSON(c.Session.MGet())
}

func (this *Admin) Lock(c *cosweb.Context) error {
	name, ok := c.Session.Get("name")
	if !ok {
		return errors.New("name empty")
	}
	<-time.After(time.Second * 10)
	return c.String(name.(string))
}

func (this *Admin) Time(c *cosweb.Context) error {
	if !c.Session.Set("time", time.Now().Unix()) {
		return errors.New("session set error")
	}
	return c.String("ok")
}
