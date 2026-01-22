package session

// Get 获取指定键的值
func (this *Data) Get(key string) any {
	return this.values.Get(key)
}

// GetString 获取指定键的字符串值
func (this *Data) GetString(key string) string {
	return this.values.GetString(key)
}

// GetInt 获取指定键的整数值
func (this *Data) GetInt(key string) int {
	return this.values.GetInt(key)
}

func (this *Data) GetInt32(key string) int32 {
	return this.values.GetInt32(key)
}

// GetInt64 获取指定键的64位整数值
func (this *Data) GetInt64(key string) int64 {
	return this.values.GetInt64(key)
}

// GetFloat64 获取指定键的浮点数值
func (this *Data) GetFloat64(key string) float64 {
	return this.values.GetFloat64(key)
}
