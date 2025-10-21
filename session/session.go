package session

import (
	"strings"

	"github.com/hwcer/cosgo/random"
)

func New(d ...*Data) *Session {
	r := &Session{}
	if len(d) > 0 {
		r.Data = d[0]
	}
	return r
}

const TokenSecretName = "_TS_"

type Session struct {
	*Data
	dirty []string
}

func (this *Session) Refresh() (string, error) {
	if this.Data == nil {
		return "", ErrorSessionNotCreate
	}
	secret := random.Strings.String(ContextRandomStringLength)
	token := strings.Join([]string{secret, this.Data.Id()}, "")

	dirty := map[string]any{}
	dirty[TokenSecretName] = secret
	this.Data.Update(dirty)
	if err := Options.Storage.Update(this.Data, dirty); err != nil {
		return "", err
	}
	return token, nil
}

// Verify 验证TOKEN信息是否有效,并初始化session
func (this *Session) Verify(token string) (err error) {
	if Options.Storage == nil {
		return ErrorStorageNotSet
	}
	if token == "" {
		return ErrorSessionTokenEmpty
	}
	if len(token) <= ContextRandomStringLength {
		return ErrorSessionIllegal
	}
	id := token[ContextRandomStringLength:]

	if this.Data, err = Options.Storage.Get(id); err != nil {
		return err
	} else if this.Data == nil {
		return ErrorSessionNotExist
	}
	secret := this.Data.GetString(TokenSecretName)
	if secret == "" {
		return ErrorSessionIllegal
	}
	if secret != token[0:ContextRandomStringLength] {
		return ErrorSessionReplaced
	}
	return nil
}

func (this *Session) Set(key string, val any) {
	if this.Data == nil {
		return
	}
	this.Data.Set(key, val)
	this.dirty = append(this.dirty, key)
}

// Update 批量修改Session信息
func (this *Session) Update(vs map[string]any) {
	if this.Data == nil {
		return
	}
	this.Data.Update(vs)
	for k, _ := range vs {
		this.dirty = append(this.dirty, k)
	}
}

func (this *Session) New(data *Data) (token string, err error) {
	if Options.Storage == nil {
		return "", ErrorStorageNotSet
	}
	if err = Options.Storage.New(data); err != nil {
		return "", err
	}
	this.Data = data
	return this.Refresh()
}

// Create 创建SESSION，uuid 用户唯一ID，可以检测是不是重复登录
func (this *Session) Create(uuid string, data map[string]any) (token string, err error) {
	if Options.Storage == nil {
		return "", ErrorStorageNotSet
	}
	if this.Data, err = Options.Storage.Create(uuid, data); err != nil {
		return "", err
	}
	return this.Refresh()
}

func (this *Session) Delete() (err error) {
	if Options.Storage == nil || this.Data == nil {
		return nil
	}
	if err = Options.Storage.Delete(this.Data); err != nil {
		return
	}
	this.release()
	return
}

// Release 释放 session 由HTTP SERVER 自动调用
func (this *Session) Release() {
	if this.Data == nil {
		return
	}
	dirty := map[string]any{}
	for _, k := range this.dirty {
		dirty[k] = this.Data.Get(k)
	}
	_ = Options.Storage.Update(this.Data, dirty)
	this.release()
}

func (this *Session) release() {
	this.dirty = nil
	this.Data = nil
}
