package utils

//MaxNumber l位所能表示的最正大数,0<l<=64
func MaxNumber(l int) uint64 {
	var v uint64 = 1
	for i := 1; i < l; i++ {
		v |= 1 << i
	}
	return v
}
