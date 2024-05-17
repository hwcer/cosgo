package random

import (
	"sort"
	"testing"
)

var items = map[int32]int32{}

func init() {
	items[1] = 5000
	items[2] = 4500
	items[3] = 490
	items[4] = 8
	items[5] = 2
}

func TestRoll(t *testing.T) {

	r := map[int32]int32{}
	rnd := New(items)

	var m int
	for i := 1; i <= 1000000; i++ {
		n := rnd.Roll()
		r[n] += 1
		if n == int32(len(items)) && m == 0 {
			m = i
		}
	}
	logf(t, r, m)
}

func TestWeight(t *testing.T) {
	rnd := New(items)
	r2 := map[int32]int32{}
	var m int
	for i := 1; i <= 1000000; i++ {
		n := rnd.Weight()
		r2[n] += 1
		if n == int32(len(items)) && m == 0 {
			m = i
		}
	}
	logf(t, r2, m)

}
func logf(t *testing.T, r map[int32]int32, m int) {
	t.Logf("首次出现极限命中发生在第%v次", m)
	var sr [][2]int32
	for k, v := range r {
		sr = append(sr, [2]int32{k, v})
	}
	sort.Slice(sr, func(i, j int) bool {
		return sr[i][0] < sr[j][0]
	})
	for _, v := range sr {
		t.Logf("key:%v,Num:%v,Weight:%v", v[0], v[1], items[v[0]])
	}
}
