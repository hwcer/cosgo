package storage

type Setter interface {
	Id() string
}

type NewSetter func(id string, val any) Setter

func NewSetterDefault(id string, data any) Setter {
	return &SetterDefault{id: id, data: data}
}

type SetterDefault struct {
	id   string
	data any
}

func (this *SetterDefault) Id() string {
	return this.id
}

func (this *SetterDefault) Get() any {
	return this.data
}

func (this *SetterDefault) Set(data any) {
	this.data = data
}
