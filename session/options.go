package session

type Options struct {
	MaxAge    int64 //有效期(S)
	MapSize   int32
	Heartbeat int32 //心跳(S)
}

type Dataset interface {
	Set(map[string]interface{})
	Get() map[string]interface{}
	Lock() bool
	UnLock()
}

type Storage interface {
	Get(string) (Dataset, bool)
	Set(key string, val map[string]interface{}) Dataset
	Del(key string)
	Close()
}
