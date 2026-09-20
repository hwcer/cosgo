package session

// Storage 存储接口。
// 🔴 Unset 为强制实现:键删除必须写穿存储(Redis 后端为 HDEL),否则下次 Verify
// 还原的副本会从存储读回旧值,内存与存储永久分叉。旧设计里这是可选接口
// (StorageDeleter),未实现时降级为写空串——两套语义并存,删除是否真正生效
// 取决于实现方是否记得多实现一个接口,坑;统一收编为接口方法,内存后端等
// "Data 与存储共享实例"的实现返回 nil 即可
type Storage interface {
	New(data *Data) error
	Get(id string) (data *Data, err error)
	Create(uuid string, value map[string]any) (data *Data, err error)
	Update(data *Data, value map[string]any) error
	Delete(data *Data) error
	// Unset 删除指定键。语义与 Session.Unset 对齐:真删除,不退化为写空串
	Unset(data *Data, keys ...string) error
}
