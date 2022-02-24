package utils

const MAXBITVALUE = 7

type Bitwise uint8

//Has bit位是否设置值
func (m *Bitwise) Has(bit int) bool {
	if bit > MAXBITVALUE {
		return false
	}
	return (*m & 1 << bit) > 0
}

//Set bit位设置为1
func (m *Bitwise) Set(bit int) {
	if bit > MAXBITVALUE {
		return
	}
	*m |= 1 << bit
}

//Delete bit位设置为0
func (m *Bitwise) Delete(bit int) {
	if !m.Has(bit) {
		return
	}
	*m -= 1 << bit
}
