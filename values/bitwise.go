package values

const MaxBitwiseUnit = 32

//Bitwise

type Byte uint32
type Bitwise []Byte

//func (b *Byte) UnmarshalJSON(b2 []byte) error {
//	*b = b2
//	return nil
//}
//
//func (b *Bitwise) MarshalJSON() ([]byte, error) {
//	if b == nil || len(*b) == 0 {
//		return nil, nil
//	}
//	v := fmt.Sprintf("%v", *b)
//	return []byte(v), nil
//}

// Has bit位是否设置值
func (m *Byte) Has(bit int) bool {
	if bit >= 0 && bit < MaxBitwiseUnit {
		return *m&(1<<bit) != 0
	}
	return false
}

// Set bit位设置为1
func (m *Byte) Set(bit int) {
	if bit >= 0 && bit < MaxBitwiseUnit {
		*m |= 1 << bit
	}
}

// Delete bit位设置为0
func (m *Byte) Delete(bit int) {
	if bit >= 0 && bit < MaxBitwiseUnit {
		*m &^= 1 << bit
	}
}

// Has bit位是否设置值
func (m *Bitwise) Has(bit int) bool {
	if bit < 0 {
		return false
	}
	i := bit / MaxBitwiseUnit
	if i >= len(*m) {
		return false
	}
	return (*m)[i]&(1<<(bit%MaxBitwiseUnit)) != 0
}

// Set bit位设置为1
func (m *Bitwise) Set(bit int) {
	if bit < 0 {
		return
	}
	i := bit / MaxBitwiseUnit
	j := bit % MaxBitwiseUnit
	b := *m
	if i >= len(b) {
		b = append(b, make(Bitwise, i+1-len(b))...)
		*m = b
	}
	b[i] |= 1 << j
}

// Delete bit位设置为0
func (m *Bitwise) Delete(bit int) {
	if bit < 0 {
		return
	}
	i := bit / MaxBitwiseUnit
	if i >= len(*m) {
		return
	}
	(*m)[i] &^= 1 << (bit % MaxBitwiseUnit)
}
