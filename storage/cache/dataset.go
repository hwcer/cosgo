package cache

type Dataset interface {
	Id() uint64
	Get() interface{}
	Set(interface{})
	Reset(id uint64, data interface{})
}

type NewDataFunc func(id uint64, data interface{}) Dataset

func NewData() *Data {
	return &Data{}
}

type Data struct {
	id   uint64
	data interface{}
}

func (this *Data) Id() uint64 {
	return this.id
}
func (this *Data) Get() interface{} {
	return this.data
}
func (this *Data) Set(data interface{}) {
	this.data = data
}

func (this *Data) Reset(id uint64, data interface{}) {
	this.id = id
	this.data = data
}
