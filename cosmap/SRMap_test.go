package cosmap

import (
	"math/rand"
	"strconv"
	"sync"
	"testing"
)

var num int = 20000
var srmap *SRMap
var syncMap sync.Map
var ArrayMap *Array

var ArrayMapKeys []ArrayKey

func init() {
	srmap = NewSRMap(num)
	syncMap = sync.Map{}
	ArrayMap = NewArray(num)
	for i := 0; i <= num; i++ {
		k := strconv.Itoa(i)
		srmap.Set(k, i)
		syncMap.Store(k, i)
		ArrayMapKeys = append(ArrayMapKeys, ArrayMap.Add(newArrayMapVal(i)))
	}
}

type arrayMapVal struct {
	id   ArrayKey
	data int
}

func (this *arrayMapVal) GetArrayKey() ArrayKey {
	return this.id
}
func (this *arrayMapVal) SetArrayKey(id ArrayKey) {
	this.id = id
}
func newArrayMapVal(v int) *arrayMapVal {
	return &arrayMapVal{data: v}
}

func Roll() string {
	i := rand.Intn(num-1) + 1
	return strconv.Itoa(i)
}

func RollArrayKey() ArrayKey {
	i := rand.Intn(num-1) + 1
	return ArrayMapKeys[i]
}
func BenchmarkSRMap(b *testing.B) {
	for i := 0; i < 10000; i++ {
		k := Roll()
		srmap.Set(k, 1)
	}
	for i := 0; i < 100000; i++ {
		k := Roll()
		srmap.Get(k)
	}

}

func BenchmarkArrayMap(b *testing.B) {
	for i := 0; i < 10000; i++ {
		RollArrayKey()
		ArrayMap.Add(newArrayMapVal(i))
	}
	for i := 0; i < 100000; i++ {
		k := RollArrayKey()
		ArrayMap.Get(k)
	}

}
func BenchmarkSYNCMap(b *testing.B) {
	for i := 0; i < 10000; i++ {
		k := Roll()
		syncMap.Store(k, 1)
	}
	for i := 0; i < 100000; i++ {
		k := Roll()
		syncMap.Load(k)
	}

}
