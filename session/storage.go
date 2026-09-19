package session

// Storage 存储接口
type Storage interface {
	New(data *Data) error
	Get(id string) (data *Data, err error)
	Create(uuid string, value map[string]any) (data *Data, err error)
	Update(data *Data, value map[string]any) error
	Delete(data *Data) error
}

// StorageDeleter 可选接口:实现后 Session.Unset 的键删除会写穿存储(Redis 后端为 HDEL)。
// 未实现时退化为把删除键写成空串——Go nil 经 go-redis 序列化为 "",与旧行为兼容。
// 🔴 没有此接口时,Session.Unset 的删除在 Redis 后端不生效:下次 Verify 还原的
// 副本会从存储读回旧值,内存与存储永久分叉。
type StorageDeleter interface {
	DeleteKeys(data *Data, keys ...string) error
}
