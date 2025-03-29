package safety

type Status int32

const (
	StatusNone    Status = 0
	StatusEnable         = 1 //白名单
	StatusDisable        = 2 //黑名单
)
