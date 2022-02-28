package registry

import "reflect"

func NewNode(name string, i reflect.Value) *Node {
	return &Node{name: name, i: i, method: make(map[string]reflect.Value)}
}

type Node struct {
	i      reflect.Value
	name   string
	method map[string]reflect.Value
}

func (this *Node) Range(prefix string, fn RegistryRangeHandle) (err error) {
	for k, m := range this.method {
		if err = fn(prefix+k, m); err != nil {
			return
		}
	}
	return nil
}
