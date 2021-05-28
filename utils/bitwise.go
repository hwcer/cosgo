package utils

type Bitwise uint64

var bitwiseSitu uint64 = 1

//位运算
func (m *Bitwise) Has(f int) bool {
	return (*m & Bitwise(bitwiseSitu<<uint64(f))) > 0
}

func (m *Bitwise) Add(f int) {
	*m |= Bitwise(bitwiseSitu << uint64(f))
}
func (m *Bitwise) Del(f int) {
	if m.Has(f) {
		*m -= Bitwise(bitwiseSitu << uint64(f))
	}
}
