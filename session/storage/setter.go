package storage

type Setter interface {
	Id() string
	Get() interface{}
	Set(interface{})
}

type NewSetter func(id string, val interface{}) Setter

// NewData 默认存储对象
func NewData(id string, data interface{}) *Data {
	return &Data{id: id, data: data}
}

func NewSetterDefault(id string, val interface{}) Setter {
	return NewData(id, val)
}

type Data struct {
	id   string
	data interface{}
}

func (this *Data) Id() string {
	return this.id
}

func (this *Data) Get() interface{} {
	return this.data
}

func (this *Data) Set(data interface{}) {
	this.data = data
}
