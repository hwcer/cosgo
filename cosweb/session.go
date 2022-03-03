package cosweb

import (
	"github.com/hwcer/cosgo/session"
	"github.com/hwcer/cosgo/session/options"
	"net/http"
	"time"
)

type SessionStartType uint8

const (
	SessionStartTypeNone SessionStartType = 0 //不需要验证登录
	SessionStartTypeAuth SessionStartType = 1 //需要登录
	SessionStartTypeLock SessionStartType = 2 //需要登录，并且锁定,用户级别防并发
)

func NewSession(c *Context) *Session {
	return &Session{c: c}
}

type Session struct {
	c      *Context
	sid    string
	key    string
	uuid   string //用户唯一标志
	cache  map[string]interface{}
	dirty  []string
	locked bool
}

func (this *Session) Start(level SessionStartType, sid ...string) (err error) {
	storage := session.Get()
	if storage == nil {
		return options.ErrorStorageNotSet
	}
	if level == SessionStartTypeNone {
		return nil
	}
	if len(sid) > 0 {
		this.sid = sid[0]
	} else {
		this.sid = this.c.Cookie.Get(session.Options.Name)
	}
	if this.sid == "" {
		return options.ErrorSessionIdEmpty
	}
	if this.key, err = options.Decode(this.sid); err != nil {
		return err
	}

	var lock bool
	var data map[string]interface{}
	if level == SessionStartTypeLock {
		lock = true
	}

	if this.uuid, data, err = storage.Get(this.key, lock); err != nil {
		return err
	} else if data == nil {
		return options.ErrorSessionNotExist
	}
	this.cache = data
	this.locked = lock
	return nil
}

func (this *Session) Get(key string) (v interface{}) {
	if this.cache != nil {
		v, _ = this.cache[key]
	}
	return
}

func (this *Session) Set(key string, val interface{}) bool {
	if this.cache == nil {
		return false
	}
	this.dirty = append(this.dirty, key)
	this.cache[key] = val
	return true
}

func (this *Session) All() map[string]interface{} {
	data := make(map[string]interface{}, len(this.cache))
	for k, v := range this.cache {
		data[k] = v
	}
	return data
}

//UUid 获取玩家uuid
func (this *Session) UUid() string {
	return this.uuid
}

func (this *Session) GetString(key string) (v string) {
	data := this.Get(key)
	if data == nil {
		return
	}
	v, _ = data.(string)
	return
}

//Create 创建SESSION，uuid 用户唯一ID，可以检测是不是重复登录
func (this *Session) Create(uuid string, data map[string]interface{}) (sid string, err error) {
	storage := session.Get()
	if storage == nil {
		return "", options.ErrorStorageNotSet
	}
	values := make(map[string]interface{}, len(data))
	for k, v := range data {
		values[k] = v
	}

	expires := this.expires()
	this.sid, this.key, err = storage.Create(uuid, values, expires, true)
	if err != nil {
		return "", err
	}
	this.uuid = uuid
	this.cache = values
	this.locked = true
	if session.Options.Name != "" {
		cookie := &http.Cookie{
			Name:  session.Options.Name,
			Value: this.sid,
			Path:  "/",
		}
		if expires > 0 {
			cookie.Expires = time.Unix(expires, 0)
		}
		this.c.Cookie.SetCookie(cookie)
	}
	return this.sid, nil
}

func (this *Session) Delete() (err error) {
	if this.sid == "" || this.key == "" {
		return nil
	}
	storage := session.Get()
	if err = storage.Delete(this.key); err != nil {
		return
	}
	this.reset()
	return
}

func (this *Session) reset() {
	this.sid = ""
	this.uuid = ""
	this.key = ""
	this.cache = nil
	this.dirty = nil
	this.locked = false
}

//release 释放 session 由HTTP SERVER
func (this *Session) release() {
	if this.sid == "" || this.key == "" || this.cache == nil {
		return
	}
	var data map[string]interface{}
	if len(this.dirty) > 0 {
		data = make(map[string]interface{})
		for _, k := range this.dirty {
			data[k] = this.cache[k]
		}
	}
	expires := this.expires()
	storage := session.Get()
	storage.Save(this.key, data, expires, this.locked)
	this.reset()
}

func (this *Session) expires() int64 {
	if session.Options.MaxAge > 0 {
		return time.Now().Add(time.Second * time.Duration(session.Options.MaxAge)).Unix()
	}
	return 0
}
