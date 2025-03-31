package storage

type Setter interface {
	Id() string
	Get() interface{}
	Set(interface{})
}

type NewSetter func(id string, val interface{}) Setter

func NewSetterDefault(id string, data interface{}) Setter {
	return &SetterDefault{id: id, data: data}
}

type SetterDefault struct {
	id   string
	data any
}

func (this *SetterDefault) Id() string {
	return this.id
}

func (this *SetterDefault) Get() interface{} {
	return this.data
}

func (this *SetterDefault) Set(data interface{}) {
	this.data = data
}
