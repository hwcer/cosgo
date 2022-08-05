package safety

type Status int32

const (
	StatusNone    Status = 0
	StatusEnable         = 1 //白名单
	StatusDisable        = 2 //黑名单
)

/*
10.0.0.0~10.255.255.255（A类）
172.16.0.0~172.31.255.255（B类）
192.168.0.0~192.168.255.255（C类）
*/
//内网IP段
var localAddress = []string{"127.0.0.1", "10.0.0.0~10.255.255.255", "172.16.0.0~172.31.255.255", "192.168.0.0~192.168.255.255"}
