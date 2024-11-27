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
	registry := New(nil)

	service := registry.Service("srv")
	_ = service.Register(&services{})

	if err := service.Register(ABC); err != nil {
		t.Logf("ERROR:%v", err)
	}

	for _, s := range registry.Services() {
		s.Range(func(node *Node) bool {
			t.Logf("route:%v", node.Route())
			return true
		})

	}

}
