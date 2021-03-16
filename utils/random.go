package utils

import (
	"math/rand"
	"time"
)

var Random *random

func init() {
	Random = NewRandom("0123456789abcdefghijklmnopqrstuvwxyz")
}

type random struct {
	rand  *rand.Rand
	bytes []byte
}

func NewRandom(randomStringSeed string) *random {
	rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	return &random{bytes: []byte(randomStringSeed), rand: rand}
}

func (this *random) Roll(a, b int) int {
	return a + this.rand.Intn(b-a)
}

func (this *random) String(l int) string {
	result := make([]byte, l)
	m := len(this.bytes) - 1
	for i := 0; i < l; i++ {
		result[i] = this.bytes[this.Roll(0, m)]
	}
	return string(result)
}
