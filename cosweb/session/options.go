package session

type Options struct {
	MaxAge    int64 //有效期(S)
	MapSize   int32
	Heartbeat int32 //心跳(S)
}

type Session interface {
	Get(string) (*Storage, bool)
	Set(key string, data map[string]interface{}) bool
	Ceate(data map[string]interface{}) *Storage
	Remove(key string) bool
	Close()
}
