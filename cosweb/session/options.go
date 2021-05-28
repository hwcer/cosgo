package session

type Options struct {
	MaxAge    int64 //有效期(S)
	MapSize   int32
	Heartbeat int32 //心跳(S)
}
