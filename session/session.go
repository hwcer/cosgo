// Package session 提供会话管理功能，支持内存和Redis存储
package session

import (
	"time"

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
func NewWithValues(uuid string, vs map[string]any) *Session {
	d := NewData(uuid, vs)
	return New(d)
}

const TokenSecretName = "_TS_"

type Session struct {
	*Data
	dirty    map[string]struct{} //修改过的键
	dirtyDel map[string]struct{} //删除过的键:与修改键分开记录,Storage 实现 StorageDeleter 时走真删除(HDEL)
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

// Unset 删除会话键。删除与修改语义不同:Storage 实现 StorageDeleter 时走真删除
// (Redis 后端为 HDEL);否则退化为写空串(与旧行为兼容)。
// 🔴 旧版本没有任何键删除写穿路径——Data.Delete 只删内存,Redis 后端下次 Verify
// 还原的副本会从存储读回旧值,删除操作等于没发生过。
func (this *Session) Unset(keys ...string) {
	if this.Data == nil || len(keys) == 0 {
		return
	}
	this.Data.Mutex(func(setter Setter) {
		for _, k := range keys {
			setter.Delete(k)
		}
	})
	this.markDirtyDel(keys...)
}

// markDirtyDel 标记删除过的键(与 markDirty 同为单请求上下文内使用,见文件头注释 #1)
func (this *Session) markDirtyDel(keys ...string) {
	if len(keys) == 0 {
		return
	}
	if this.dirtyDel == nil {
		this.dirtyDel = make(map[string]struct{}, len(keys))
	}
	for _, k := range keys {
		this.dirtyDel[k] = struct{}{}
	}
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
	//秘钥立即写穿存储:Redis 后端下只标脏的话,登录响应先于 Release 到达客户端,
	//第二个请求 Verify 从存储读不到秘钥——偶发"刚登录就被踢"
	if err = this.Submit(); err != nil {
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
	//同 New:秘钥立即写穿,消除并发请求窗口
	if err = this.Submit(); err != nil {
		return "", err
	}
	Emit(EventSessionCreated, this.Data)
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

// Submit 提交所有修改，不会立即release影响后续登录判断。
// 成功后清空 dirty——随后的 Release 不会把同一批键重复写一遍;
// 失败则保留,留待下一次 Submit/Release 兜底重试
func (this *Session) Submit() (err error) {
	if this.Data == nil {
		return
	}
	//删除键优先处理:走 StorageDeleter 真删除,未实现则退化为写空串
	if len(this.dirtyDel) > 0 {
		keys := make([]string, 0, len(this.dirtyDel))
		for k := range this.dirtyDel {
			keys = append(keys, k)
		}
		if sd, ok := Options.Storage.(StorageDeleter); ok {
			if err = sd.DeleteKeys(this.Data, keys...); err == nil {
				this.dirtyDel = nil
			} else {
				return
			}
		} else {
			this.markDirty(keys...)
			this.dirtyDel = nil
		}
	}
	if len(this.dirty) == 0 {
		return
	}
	dirty := map[string]any{}
	for k := range this.dirty {
		dirty[k] = this.Data.Get(k)
	}
	if len(dirty) == 0 {
		return
	}
	if err = Options.Storage.Update(this.Data, dirty); err == nil {
		this.dirty = nil
	}
	return
}

// Release 释放 session 由HTTP SERVER 自动调用。
// Submit 失败时做短暂同步重试,只覆盖瞬时抖动;刻意不做跨请求异步重试——
// 请求结束后迟到写入的旧值会回滚后续请求已落库的新值,风险大于收益。
// 持续失败以 Alert 留痕(只记键名不含值,避免泄露秘钥类字段),脏标记随
// release 清空,等同旧版行为;差异是瞬时故障已在本请求内消化。
func (this *Session) Release() {
	var err error
	for i := 0; i < 3; i++ {
		if err = this.Submit(); err == nil {
			break
		}
		time.Sleep(time.Duration(50*(i+1)) * time.Millisecond)
	}
	if err != nil {
		logger.Alert("session Submit error after retry, dropped keys: %v%v", this.dirty, this.dirtyDel)
	}
	this.release()
}

func (this *Session) release() {
	this.dirty = nil
	this.dirtyDel = nil
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
