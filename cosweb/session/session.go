package session

import (
	"github.com/hwcer/cosgo/utils"
	"net/http"
	"time"
)

const SessionContextRandomStringLength = 4

func NewSession(c Context) *Session {
	return &Session{c: c}
}

type Session struct {
	c       Context
	sid     string
	dirty   []string
	locked  bool
	dataset map[string]interface{}
}

func (this *Session) lock() bool {
	if this.sid == "" {
		return false
	}
	if Options.Storage.Lock(this.sid) {
		this.locked = true
		return true
	} else {
		return false
	}
}

func (this *Session) unlock() bool {
	if this.sid == "" || !this.locked {
		return false
	}
	defer func() {
		this.locked = false
	}()
	return Options.Storage.UnLock(this.sid)
}

func (this *Session) Start(level int) (err error) {
	if Options.Storage == nil {
		return ErrorStorageNotSet
	}
	if level < 1 {
		return nil
	}
	var sid string
	var dataset map[string]interface{}
	if sid, err = this.decode(); err != nil {
		return err
	}
	if dataset, err = Options.Storage.Get(sid); err != nil {
		return err
	} else if dataset == nil {
		return ErrorSessionNotExist
	}
	this.dataset = dataset
	if level > 1 && !this.lock() {
		return ErrorSessionLocked
	}
	return nil
}

func (this *Session) Get(key string) (interface{}, bool) {
	if this.dataset == nil {
		return nil, false
	}
	if v, ok := this.dataset[key]; ok {
		return v, true
	} else {
		return nil, false
	}
}

func (this *Session) Set(key string, val interface{}) bool {
	if this.dataset == nil {
		return false
	}
	this.dirty = append(this.dirty, key)
	this.dataset[key] = val
	return true
}

//MGet 获取所有SESSION
func (this *Session) MGet() map[string]interface{} {
	return this.dataset
}

func (this *Session) Create(data map[string]interface{}) (string, error) {
	if Options.Storage == nil {
		return "", ErrorStorageNotSet
	}
	this.locked = true
	id := Options.Storage.Create(data)
	sid, err := this.encode(id)
	if err != nil {
		return "", err
	}
	this.sid = sid
	this.dataset = data
	if Options.Name != "" {
		cookie := &http.Cookie{
			Name:  Options.Name,
			Value: sid,
		}
		var ttl int64
		if ttl, err = Options.Storage.TTL(id); err != nil {
			return "", err
		} else if ttl > 0 {
			cookie.Expires = time.Unix(ttl, 0)
		}
		this.c.SetCookie(cookie)
	}
	return sid, nil
}

//Reset 设置 session id 可能是通过非COOKIE的其他方式传递sid
func (this *Session) Reset(sid string) {
	this.sid = sid
}

//Release 释放 session 由HTTP SERVER
func (this *Session) Release() {
	if this.sid == "" || this.dataset == nil {
		return
	}
	if len(this.dirty) > 0 {
		data := make(map[string]interface{})
		for _, k := range this.dirty {
			data[k] = this.dataset[k]
		}
		Options.Storage.Set(this.sid, data)
	}
	this.unlock()
	Options.Storage.Expire(this.sid)
	this.sid = ""
	this.dataset = nil
}

func (this *Session) decode() (string, error) {
	if this.sid == "" {
		cookie, err := this.c.GetCookie(Options.Name)
		if err == nil && cookie != nil {
			this.sid = cookie.Value
		}
	}

	if this.sid == "" {
		return "", ErrorSessionIdEmpty
	}
	if Options.Secret == "" {
		return this.sid, nil
	}
	str, err := utils.Crypto.AESDecrypt(this.sid, Options.Secret)
	if err != nil {
		return "", err
	}
	//fmt.Printf("%v--%v\n", str, len(str))
	return str[SessionContextRandomStringLength:], nil
}
func (this *Session) encode(key string) (string, error) {
	if Options.Secret == "" {
		return key, nil
	}
	s := utils.Random.String(SessionContextRandomStringLength)
	//fmt.Printf("%v--%v---%v\n", key, s, s+key)
	//fmt.Printf("%v--%v---%v\n", len(key), len(s), len(s+key))
	return utils.Crypto.AESEncrypt(s+key, Options.Secret)
}
