package session

// ContextRandomStringLength Token 前缀随机串长度
const ContextRandomStringLength = 6

// Options 全局配置
var Options = struct {
	Name      string  // session cookie name
	MaxAge    int64   // 有效期（秒），0 = 不过期
	Storage   Storage // 存储后端
	Heartbeat int32   // 心跳间隔（秒），用于内存后端自动清理过期会话
}{
	Name:      "_cookie_vars",
	MaxAge:    3600,
	Heartbeat: 10,
}
