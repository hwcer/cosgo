package registry

import "strings"

func NewPrefix(name string) *Prefix {
	p := &Prefix{
		name: Format(name),
	}
	p.index = len(p.name)
	return p
}

type Prefix struct {
	name  string //前缀
	index int    //前缀长度
}

func (this *Prefix) Name() string {
	return this.name
}

func (this *Prefix) Index() int {
	return this.index
}

func (this *Prefix) Format(path string) (r string) {
	r = Format(path)
	if this.name !="/"{
		r = strings.TrimPrefix(path,this.name)
	}
	return
}
