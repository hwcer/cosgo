package cosnet

import (
	"testing"
)

func TestMessage(t *testing.T) {
	var data []byte
	t.Logf("data nil len:%v", len(data))
	data = make([]byte, 2)
	t.Logf("data make len:%v", len(data))
}
