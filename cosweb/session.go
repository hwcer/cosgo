package cosweb

import (
	"errors"
	"github.com/hwcer/cosgo/cosweb/session"
	"github.com/hwcer/cosgo/utils"
	"net/http"
)

type Session struct {
	Key     string
	Method  []int //存放SESSION KEY的方式
	Secret  string
	Storage session.Session //Session数据存储器
}

func NewSession() *Session {
	return &Session{
		Key:    "CosWebSessId",
		Method: []int{RequestDataTypeCookie, RequestDataTypeQuery},
		Secret: utils.Random.String(16),
	}
}

const sessionRandomStringLength = 4

type SessionContext struct {
	c      *Context
	data   map[string]interface{}
	store  *session.Storage
	update []string
	locked bool
}

func NewSessionContext() *SessionContext {
	return &SessionContext{}
}

func (this *SessionContext) reset() {

}
func (this *SessionContext) release() {
	this.finish()
	this.data = nil
	this.update = nil
	this.locked = false
}

func (this *SessionContext) Start(c *Context, level int) error {
	if this.c.Server.Options.Session.Storage == nil {
		return errors.New("Server Session Storage is nil")
	}
	this.c = c
	if level <= 0 {
		return nil
	}
	key, err := this.decode()
	if err != nil {
		return err
	}
	store, ok := this.c.Server.Options.Session.Storage.Get(key)
	if !ok {
		return errors.New("session not exist")
	} else if store == nil {
		return errors.New("session expired")
	}
	this.store = store
	if level == 1 {
		return nil
	}
	if !this.store.Lock() {
		return errors.New("session Locked")
	}
	this.locked = true
	return nil
}

func (this *SessionContext) Get(key string) (interface{}, bool) {
	if this.store == nil {
		return nil, false
	}
	if this.data == nil {
		this.data = this.store.Get()
	}
	v, ok := this.data[key]
	return v, ok
}
func (this *SessionContext) Set(key string, val interface{}) bool {
	if this.store == nil {
		return false
	}
	this.data[key] = val
	this.update = append(this.update, key)
	return true
}

func (this *SessionContext) Create(key string, val map[string]interface{}) (string, error) {
	if this.c.Server.Options.Session.Storage == nil {
		return "", errors.New("Server Session Storage is nil")
	}

	this.data = val
	this.store = this.c.Server.Options.Session.Storage.Ceate(val)
	this.locked = true

	sid, err := this.encode(this.store.GetStorageKey())
	if err != nil {
		return "", err
	}

	if utils.IndexOf(this.c.Server.Options.Session.Method, RequestDataTypeCookie) >= 0 {
		this.c.SetCookie(&http.Cookie{Name: this.c.Server.Options.Session.Key, Value: sid})
	}
	return sid, nil
}

func (this *SessionContext) finish() {
	if len(this.update) > 0 {
		d := make(map[string]interface{}, len(this.update))
		for k, v := range this.data {
			d[k] = v
		}
		this.store.Set(d)
	}
	if this.locked {
		this.store.UnLock()
	}
}

func (this *SessionContext) decode() (string, error) {
	sid := this.c.Get(this.c.Server.Options.Session.Key, this.c.Server.Options.Session.Method...)
	if sid == "" {
		return "", errors.New("sid empty")
	}
	if this.c.Server.Options.Session.Secret == "" {
		return sid, nil
	}
	str, err := utils.Crypto.AESDecrypt(sid, this.c.Server.Options.Session.Secret)
	if err != nil {
		return "", err
	}
	//fmt.Printf("%v--%v\n", str, len(str))
	return str[sessionRandomStringLength:], nil
}
func (this *SessionContext) encode(key string) (string, error) {
	if this.c.Server.Options.Session.Secret == "" {
		return key, nil
	}
	s := utils.Random.String(sessionRandomStringLength)
	//fmt.Printf("%v--%v---%v\n", key, s, s+key)
	//fmt.Printf("%v--%v---%v\n", len(key), len(s), len(s+key))
	return utils.Crypto.AESEncrypt(s+key, this.c.Server.Options.Session.Secret)
}
