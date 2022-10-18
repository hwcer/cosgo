package binder

import (
	"testing"
)

func TestBinder(t *testing.T) {
	bind := New(MIMEJSON)
	var a int32 = 100
	b, err := bind.Marshal(a)
	if err != nil {
		t.Logf("Marshal Error:%v", err)
	} else {
		t.Logf("Marshal Success:%v", b)
	}

	var c int32

	err = bind.Unmarshal(b, &c)
	if err != nil {
		t.Logf("Unmarshal Error:%v", err)
	} else {
		t.Logf("Unmarshal Success:%v", c)
	}

}
