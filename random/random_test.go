package random

import (
	"sort"
	"testing"
)

var items = map[int32]int32{}

func TestName(t *testing.T) {
	items[1] = 8000
	items[2] = 1000
	items[3] = 999
	items[4] = 1

	r := map[int32]int32{}
	rnd := New(items)
	for i := 0; i < 100000; i++ {
		n := rnd.Roll()
		r[n] += 1
	}
	logf(t, r, "Roll")

	t.Logf("-----------Weight--------------")
	r2 := map[int32]int32{}
	for i := 0; i < 100000; i++ {
		n := rnd.Weight()
		r2[n] += 1
	}
	logf(t, r2, "Weight")

}

func logf(t *testing.T, r map[int32]int32, name string) {
	var sr [][2]int32
	for k, v := range r {
		sr = append(sr, [2]int32{k, v})
	}
	sort.Slice(sr, func(i, j int) bool {
		return sr[i][0] < sr[j][0]
	})
	for _, v := range sr {
		t.Logf("%v key:%v,Num:%v,Weight:%v", name, v[0], v[1], items[v[0]])
	}
}
