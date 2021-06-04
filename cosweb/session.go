package cosweb

import (
	"errors"
	"github.com/hwcer/cosgo/storage"
	"github.com/hwcer/cosgo/utils"
	"net/http"
	"time"
)

const SessionContextRandomStringLength = 4

func NewSessionContext(c *Context) *SessionContext {
	return &SessionContext{c: c}
}

type sessionOptions struct {
	Name   string
	Secret string
	Method RequestDataTypeMap
}

type SessionContext struct {
	c       *Context
	locked  bool
	dataset storage.Dataset
}

func (this *SessionContext) Start(level int) error {
	if level < 1 {
		return nil
	}
	sid, err := this.decode()
	if err != nil {
		return err
	}
	dataset, ok := this.c.engine.Storage.Get(sid)
	if !ok {
		return errors.New("session not exist")
	} else if dataset == nil {
		return errors.New("session expired")
	}
	this.dataset = dataset
	if level == 1 {
		return nil
	}
	if !this.dataset.Lock() {
		return errors.New("session Locked")
	}
	this.locked = true
	return nil
}

func (this *SessionContext) Get(key string) (interface{}, bool) {
	if this.dataset == nil {
		return nil, false
	}
	return this.dataset.Get(key)
}
func (this *SessionContext) Set(key string, val interface{}) bool {
	if this.dataset == nil {
		return false
	}
	this.dataset.Set(key, val)
	return true
}

func (this *SessionContext) Create(val map[string]interface{}) (string, error) {
	this.dataset = this.c.engine.Storage.Create(val)
	this.locked = true
	sid, err := this.encode(this.dataset.Id())
	if err != nil {
		return "", err
	}
	if Options.Session.Method.IndexOf(RequestDataTypeCookie) >= 0 {
		cookie := &http.Cookie{
			Name:  Options.Session.Name,
			Value: sid,
		}
		if expire := this.dataset.Expire(); expire > 0 {
			cookie.Expires = time.Unix(expire, 0)
		}
		this.c.SetCookie(cookie)
	}
	return sid, nil
}

func (this *SessionContext) Close() {
	if this.dataset == nil {
		return
	}
	this.dataset.Reset(this.locked)
	this.locked = false
	this.dataset = nil
}

func (this *SessionContext) decode() (string, error) {
	sid := this.c.Get(Options.Session.Name, Options.Session.Method...)
	if sid == "" {
		return "", errors.New("sid empty")
	}
	if Options.Session.Secret == "" {
		return sid, nil
	}
	str, err := utils.Crypto.AESDecrypt(sid, Options.Session.Secret)
	if err != nil {
		return "", err
	}
	//fmt.Printf("%v--%v\n", str, len(str))
	return str[SessionContextRandomStringLength:], nil
}
func (this *SessionContext) encode(key string) (string, error) {
	if Options.Session.Secret == "" {
		return key, nil
	}
	s := utils.Random.String(SessionContextRandomStringLength)
	//fmt.Printf("%v--%v---%v\n", key, s, s+key)
	//fmt.Printf("%v--%v---%v\n", len(key), len(s), len(s+key))
	return utils.Crypto.AESEncrypt(s+key, Options.Session.Secret)
}
