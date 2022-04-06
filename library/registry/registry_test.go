package registry

import (
	"testing"
)

type services struct {
}

func (s *services) Test() {

}

func ABC() {

}

func TestRoute(t *testing.T) {
	registry := New("/:x")

	route := registry.Namespace("srv")
	route.Register(&services{})

	if err := route.Register(ABC); err != nil {
		t.Logf("ERROR:%v", err)
	}

	registry.Range(func(name string, method interface{}) error {
		t.Logf("NAME:%v", name)
		return nil
	})

	t.Logf("Paths:%v", registry.Paths())

}
