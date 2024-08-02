package random

import (
	"math/rand"
	"time"
)

var Strings = NewBytes([]byte("0123456789abcdefghijklmnopqrstuvwxyz"))

type Bytes struct {
	rand  *rand.Rand
	bytes []byte
}

func NewBytes(bytes []byte) *Bytes {
	return &Bytes{bytes: bytes, rand: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (this *Bytes) New(l int) []byte {
	r := make([]byte, l)
	m := int32(len(this.bytes) - 1)
	for i := 0; i < l; i++ {
		k := Roll(0, m)
		r[i] = this.bytes[k]
	}
	return r
}

func (this *Bytes) String(l int) string {
	return string(this.New(l))
}
