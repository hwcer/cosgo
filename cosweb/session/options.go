package session

var Options = struct {
	MaxAge    int64 //有效期(S)
	MapSize   int32
	Heartbeat int32 //心跳(S)
}{
	MaxAge:    3600,
	MapSize:   1024,
	Heartbeat: 10,
}

type Dataset interface {
	Id() string
	Set(key string, val interface{})
	Get(key string) (interface{}, bool)
	Lock() bool
	UnLock()
	Reset() //自动续约
	Expire() int64
}

type Storage interface {
	Get(string) (Dataset, bool)
	Start()
	Close()
	Create(map[string]interface{}) Dataset
	Remove(string) bool
}
