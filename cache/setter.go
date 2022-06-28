package cache

type Interface interface {
	Id() uint64
	Get() interface{}
	Set(interface{})
	Reset(id uint64, data interface{})
}

func NewSetter(id uint64, data interface{}) *Setter {
	return &Setter{id: id, data: data}
}

type Setter struct {
	id   uint64
	data interface{}
}

func (this *Setter) Id() uint64 {
	return this.id
}
func (this *Setter) Get() interface{} {
	return this.data
}
func (this *Setter) Set(data interface{}) {
	this.data = data
}

func (this *Setter) Reset(id uint64, data interface{}) {
	this.id = id
	this.data = data
}
