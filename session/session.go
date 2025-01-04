package session

func New() *Session {
	return &Session{}
}

type Session struct {
	*Data
	dirty []string
}

// Verify 验证TOKEN信息是否有效,并初始化session
func (this *Session) Verify(token string) (err error) {
	if Options.Storage == nil {
		return ErrorStorageNotSet
	}
	if token == "" {
		return ErrorSessionIdEmpty
	}
	if this.Data, err = Options.Storage.Verify(token); err != nil {
		return err
	} else if this.Data == nil {
		return ErrorSessionNotExist
	}
	if this.token != token {
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

// Create 创建SESSION，uuid 用户唯一ID，可以检测是不是重复登录
func (this *Session) Create(uuid string, data map[string]any) (token string, err error) {
	if Options.Storage == nil {
		return "", ErrorStorageNotSet
	}
	if this.Data, err = Options.Storage.Create(uuid, data, Options.MaxAge); err == nil {
		token = this.Data.token
	}
	return
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
	_ = Options.Storage.Update(this.Data, dirty, Options.MaxAge)
	this.release()
}

func (this *Session) release() {
	this.dirty = nil
	this.Data = nil
}
