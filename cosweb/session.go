package cosweb

import (
	"errors"
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosgo/utils"
	"net/http"
)

const sessionRandomStringLength = 4

type Session struct {
	s *Server
	c *Context

	data   map[string]interface{}
	store  session.Dataset
	update []string
	locked bool
}

func NewSession(s *Server, c *Context) *Session {
	return &Session{s: s, c: c}
}

func (this *Session) reset() {

}
func (this *Session) release() {
	this.finish()
	this.data = nil
	this.store = nil
	this.update = nil
	this.locked = false
}

func (this *Session) Start(level int) error {
	if this.s.Options.SessionStorage == nil {
		return errors.New("Server SessionStorage is nil")
	}
	if level <= 0 {
		return nil
	}
	key, err := this.decode()
	if err != nil {
		return err
	}
	data, ok := this.s.Options.SessionStorage.Get(key)
	if !ok {
		return errors.New("session not exist")
	} else if data == nil {
		return errors.New("session expired")
	}
	this.store = data
	if level == 1 {
		return nil
	}
	if !this.store.Lock() {
		return errors.New("session Locked")
	}
	this.locked = true
	return nil
}

func (this *Session) Get(key string) (interface{}, bool) {
	if this.store == nil {
		return nil, false
	}
	if this.data == nil {
		this.data = this.store.Get()
	}
	v, ok := this.data[key]
	return v, ok
}
func (this *Session) Set(key string, val interface{}) bool {
	if this.store == nil {
		return false
	}
	this.data[key] = val
	this.update = append(this.update, key)
	return true
}

func (this *Session) New(key string, val map[string]interface{}) (string, error) {
	if this.s.Options.SessionStorage == nil {
		return "", errors.New("Server SessionStorage is nil")
	}
	sid, err := this.encode(key)
	if err != nil {
		return "", err
	}
	this.store = this.s.Options.SessionStorage.Set(key, val)
	this.locked = true
	if utils.IndexOf(this.s.Options.SessionType, RequestDataTypeCookie) >= 0 {
		this.c.SetCookie(&http.Cookie{Name: this.s.Options.SessionKey, Value: sid})
	}
	return sid, nil
}

func (this *Session) finish() {
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

func (this *Session) decode() (string, error) {
	sid := this.c.Get(this.s.Options.SessionKey, this.s.Options.SessionType...)
	if sid == "" {
		return "", errors.New("sid empty")
	}
	str, err := utils.Crypto.AESDecrypt(sid, this.s.Options.SessionSecret)
	if err != nil {
		return "", err
	}
	//fmt.Printf("%v--%v\n", str, len(str))
	return str[sessionRandomStringLength:], nil
}
func (this *Session) encode(key string) (string, error) {
	s := utils.Random.String(sessionRandomStringLength)
	//fmt.Printf("%v--%v---%v\n", key, s, s+key)
	//fmt.Printf("%v--%v---%v\n", len(key), len(s), len(s+key))
	return utils.Crypto.AESEncrypt(s+key, this.s.Options.SessionSecret)
}
