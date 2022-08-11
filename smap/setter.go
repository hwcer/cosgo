package smap

type MID string

type Setter interface {
	Id() MID
	Get() interface{}
	Set(interface{})
}

func NewData(id MID, data interface{}) *Data {
	return &Data{id: id, data: data}
}

func NewSetter(id MID, val interface{}) Setter {
	return NewData(id, val)
}

type Data struct {
	id   MID
	data interface{}
}

func (this *Data) Id() MID {
	return this.id
}

func (this *Data) Get() interface{} {
	return this.data
}

func (this *Data) Set(data interface{}) {
	this.data = data
}
