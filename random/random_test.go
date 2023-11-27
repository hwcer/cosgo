package random

import "testing"

func TestName(t *testing.T) {
	items := map[int32]int32{}
	for i := int32(1); i < 10; i++ {
		items[i] = 100
	}

	r := map[int32]int32{}
	rnd := New(items)
	for i := 0; i < 10000; i++ {
		n := rnd.Roll()
		r[n] += 1
	}
	///////////////////////
	for k, v := range r {
		t.Logf("key:%v,Num:%v", k, v)
	}

	for i := 0; i < 10; i++ {
		n := rnd.Multi(2)
		t.Logf("Multi:%v   %v", i, n)
	}

}
