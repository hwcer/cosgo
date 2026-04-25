// Package session 提供会话管理功能，支持内存和Redis存储
package session

import (
	"github.com/hwcer/cosgo/random"
	"github.com/hwcer/logger"
)

// 注意：
// 1. 一个 Session 绑定的是一个用户的单次请求的上下文，不会存在并发问题
// 2. 业务层面会限制用户的并发请求以保证数据安全
// 3. dirty 字段用于记录修改过的键，在 Release 时批量更新到存储

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
	dirty map[string]struct{}
}

func (this *Session) Refresh() (string, error) {
	if this.Data == nil {
		return "", ErrorSessionNotCreate
	}
	secret := random.Strings.String(ContextRandomStringLength)
	token := secret + this.Data.Id()

	this.Data.Set(TokenSecretName, secret, func() {
		this.markDirty(TokenSecretName)
	})
	return token, nil
}

// Token 获取当前TOKEN，可能为空
func (this *Session) Token() (string, error) {
	if this.Data == nil {
		return "", ErrorSessionNotCreate
	}
	secret := this.Data.GetString(TokenSecretName)
	if secret == "" {
		return this.Refresh()
	}
	return secret + this.Data.Id(), nil
}

// Verify 验证TOKEN信息是否有效,并初始化session
func (this *Session) Verify(token string) (err error) {
	if Options.Storage == nil {
		return ErrorStorageEmpty
	}
	if token == "" {
		return ErrorSessionEmpty
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
	// 常量时间比较，防时序侧信道探测 secret 前缀
	// 直接比较字符串字节，零堆分配（避免 []byte 转换的 2 次分配）
	if !constantTimeStringEqual(secret, token[:ContextRandomStringLength]) {
		return ErrorSessionReplaced
	}
	return nil
}

func (this *Session) Set(key string, val any) {
	if this.Data == nil {
		return
	}
	this.Data.Set(key, val, func() {
		this.markDirty(key)
	})
}

// markDirty 标记修改过的键。
// Session 绑定单次请求上下文(参见文件头注释 #1),dirty 字段不会被并发读写,
// 直接 mutate 即可;原先的 Copy-on-Write 在此语义下是无意义的额外分配。
func (this *Session) markDirty(keys ...string) {
	if len(keys) == 0 {
		return
	}
	if this.dirty == nil {
		this.dirty = make(map[string]struct{}, len(keys))
	}
	for _, k := range keys {
		this.dirty[k] = struct{}{}
	}
}

func (this *Session) Update(vs map[string]any) {
	if this.Data == nil {
		return
	}
	this.Data.Update(vs, func() {
		if this.dirty == nil {
			this.dirty = make(map[string]struct{}, len(vs))
		}
		for k := range vs {
			this.dirty[k] = struct{}{}
		}
	})
}

func (this *Session) New(data *Data) (token string, err error) {
	if Options.Storage == nil {
		return "", ErrorStorageEmpty
	}
	if err = Options.Storage.New(data); err != nil {
		return "", err
	}
	this.Data = data
	if token, err = this.Refresh(); err != nil {
		return "", err
	}
	Emit(EventSessionNew, data)
	return
}

// Create 创建SESSION，uuid 用户唯一ID，可以检测是不是重复登录
func (this *Session) Create(uuid string, data map[string]any) (token string, err error) {
	if Options.Storage == nil {
		return "", ErrorStorageEmpty
	}
	this.Data, err = Options.Storage.Create(uuid, data)
	if err != nil {
		return "", err
	}
	if token, err = this.Refresh(); err != nil {
		return "", err
	}
	Emit(EventSessionCreated, data)
	return
}

func (this *Session) Delete() (err error) {
	if Options.Storage == nil || this.Data == nil {
		return nil
	}
	data := this.Data
	if err = Options.Storage.Delete(data); err != nil {
		return
	}
	this.release()
	Emit(EventSessionRelease, data)
	return
}

// Release 释放 session 由HTTP SERVER 自动调用
func (this *Session) Release() {
	if this.Data == nil || len(this.dirty) == 0 {
		this.release()
		return
	}
	dirty := map[string]any{}
	for k := range this.dirty {
		dirty[k] = this.Data.Get(k)
	}
	if len(dirty) == 0 {
		this.release()
		return
	}
	if err := Options.Storage.Update(this.Data, dirty); err != nil {
		logger.Alert("session update error: %v", err)
	}
	this.release()
}

func (this *Session) release() {
	this.dirty = nil
	this.Data = nil
}

// constantTimeStringEqual 常量时间字符串比较，零堆分配
// 功能等价于 subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
// 但避免了 string → []byte 转换产生的 2 次堆分配
func constantTimeStringEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}
