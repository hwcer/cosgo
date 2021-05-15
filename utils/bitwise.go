package utils

type Bitwise uint

var bitwiseSitu uint = 1

//位运算
func (m *Bitwise) Has(f int32) bool {
	return (*m & Bitwise(bitwiseSitu<<uint(f))) > 0
}

func (m *Bitwise) Add(f int32) {
	*m |= Bitwise(bitwiseSitu << uint(f))
}
func (m *Bitwise) Del(f int32) {
	if m.Has(f) {
		*m -= Bitwise(bitwiseSitu << uint(f))
	}
}
