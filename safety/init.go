package safety

// Status 规则状态
type Status int32

const (
	StatusNone    Status = 0 // 未匹配
	StatusEnable  Status = 1 // 白名单
	StatusDisable Status = 2 // 黑名单
)
