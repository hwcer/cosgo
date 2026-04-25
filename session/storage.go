package session

// Storage 存储接口
type Storage interface {
	New(data *Data) error
	Get(id string) (data *Data, err error)
	Create(uuid string, value map[string]any) (data *Data, err error)
	Update(data *Data, value map[string]any) error
	Delete(data *Data) error
}
