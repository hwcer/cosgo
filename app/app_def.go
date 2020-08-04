package app

var (
	// 版本信息 参数有编译脚本传入
	DINF_VER  = "unknown" // 版本号
	DINF_SRC  = "unknown" // 源码分支及版本信息
	DINF_BTM  = "unknown" // 编译时间
	DINF_GO   = "unknown" // golang版本
	DINF_HOST = "unknown" // host info

	defaultCtrl = new(appCtrl)
)

type AppBaseConf struct {
	// pid
	Id   string `json:"-" gorm:"primary_key"`
	Desc string `json:"-" gorm:"index;not null"`

	Pid string `json:"pid" gorm:"-"` // 服务器唯一索引配置 例如 1

	// net
	Mode    string `json:"mode" gorm:"size:16;not null"` // app模式 rpcd | simple | app
	RpcNet  string `json:"rpcnet"`                       // rpc网络类型 tcp udp kcp等，
	RpcAddr string `json:"rpcaddr"`                      // rpc服务监听地址 ip为内网地址， disable不启用，auto自动获取
	WanAddr string `json:"wanaddr"`                      // 公网服务地址 ip为 0.0.0.0，disable不启用， auto自动获取

	// log
	LogLv       int    `json:"loglv" gorm:"default:10;not null"`       // 日志等级
	LogMode     int    `json:"logmode" gorm:"default:1;not null"`      // 日志模式 标准输出|文件|网络
	LogPath     string `json:"logpath"`                                // 日志路径 默认工作路径 log目录
	LogOssPath  string `json:"logosspath"`                             // OSS日志路径 默认工作路径 logoss目录
	LogMaxSize  int    `json:"logmaxsize" gorm:"default:500;not null"` // 最大分片大小 单位M 默认500M
	LogInterval int    `json:"loginterval" gorm:"default:1;not null"`  // 日志命名间隔时间 默认按天
	LogUdpAddr  string `json:"logudpaddr"`                             // filebeat 日志采集服 udp地址
	LogTcpAddr  string `json:"logtcpaddr"`                             // filebeat 日志采集服 tcp地址

	// profile
	CpuNum   int    `json:"cpunum"`                                  // CPU数量设置 默认0 全部
	ProfAddr string `json:"profaddr" gorm:"default:'auto';not null"` // 性能分析工具地址  disable不启用 auto自动获取
	StatAddr string `json:"stataddr" gorm:"default:'auto';not null"` // 状态参数获取地址  disable不启用 auto自动获取

	Ver int `json:"ver"`
}

type AppBaseFlag struct {
	Console              bool   `arg:"-c, help:show log in console"`        // 日志是否输入出到标准输出
	Version              bool   `arg:"-v, help:show version info and exit"` // 显示版本信息
	Provider             string `arg:"-p, help:conf provider"`              // 配置方式 本地|配置中心
	Shard                string `arg:"-s, help:shard name"`                 // 指定shard
	ShardIdx             int    `arg:"-i, help:shard index"`                // 指定shard索引编号
	SrvIdx               int    `arg:"-I, help:srv index"`                  // 指定服务进程索引编号
	ErrorWechatNotNotify bool   `arg:"-w, help:报错消息微信通知"`             // 指定服务进程索引编号
}

func (a *AppBaseFlag) GetBase() *AppBaseFlag {
	return a
}
