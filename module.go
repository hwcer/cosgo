package cosgo

// Module 模块接口，所有需要被cosgo管理的模块都需要实现此接口
type Module interface {
	// Id 返回模块的唯一标识
	Id() string
	// Init 模块初始化，在应用启动时调用
	// 此阶段主要进行模块的初始化工作，如配置加载、资源分配等
	Init() error
	// Start 模块启动，在所有模块初始化完成后调用
	// 此阶段主要进行模块的业务逻辑启动，如启动服务、监听端口等
	Start() error
	// Close 模块关闭，在应用关闭时调用
	// 此阶段主要进行模块的资源释放、连接关闭等清理工作
	Close() error
}
