package registry

// Param 单个路由参数
type Param struct {
	Key   string
	Value string
}

// Params 路由参数切片，替代 map[string]string
// 路由参数通常 1~3 个，线性查找比 map hash 更快，且内存连续、缓存友好
type Params []Param

// Get 按名查找参数值
func (ps Params) Get(name string) (string, bool) {
	for i := range ps {
		if ps[i].Key == name {
			return ps[i].Value, true
		}
	}
	return "", false
}

// ByName 按名查找参数值，未找到返回空字符串
func (ps Params) ByName(name string) string {
	v, _ := ps.Get(name)
	return v
}
