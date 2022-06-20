package values

const MaxBitwiseUnit = 8

type Bitwise uint8
type BitSet []Bitwise

//Has bit位是否设置值
func (m *Bitwise) Has(bit int) (r bool) {
	if bit < MaxBitwiseUnit {
		r = *m&(1<<bit) > 0
	}
	return
}

//Set bit位设置为1
func (m *Bitwise) Set(bit int) {
	if bit < MaxBitwiseUnit {
		*m |= 1 << bit
	}
}

//Delete bit位设置为0
func (m *Bitwise) Delete(bit int) {
	if m.Has(bit) {
		*m -= 1 << bit
	}
}

//Has bit位是否设置值
func (m *BitSet) Has(bit int) bool {
	i := bit / MaxBitwiseUnit
	if i >= len(*m) {
		return false
	}
	j := bit % MaxBitwiseUnit
	return (*m)[i].Has(j)
}

//Set bit位设置为1
func (m *BitSet) Set(bit int) {
	b := *m
	i := bit / MaxBitwiseUnit
	j := bit % MaxBitwiseUnit
	if i >= len(b) {
		c := i + 1
		v := make(BitSet, c, c)
		copy(v, b)
		b = v
	}
	v := b[i]
	v.Set(j)
	b[i] = v
	*m = b
}

//Delete bit位设置为0
func (m *BitSet) Delete(bit int) {
	i := bit / MaxBitwiseUnit
	if i >= len(*m) {
		return
	}
	j := bit % MaxBitwiseUnit
	(*m)[i].Delete(j)
}
