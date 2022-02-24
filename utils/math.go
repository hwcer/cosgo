package utils

var Math = &math{}

type math struct {
}

func (this *math) MaxInt32(args ...int32) int32 {
	var r int32
	for _, v := range args {
		if v > r {
			r = v
		}
	}
	return r
}
func (this *math) MinInt32(args ...int32) int32 {
	var r int32
	for _, v := range args {
		if v == 0 || v < r {
			r = v
		}
	}
	return r
}

func (this *math) MaxInt64(args ...int64) int64 {
	var r int64
	for _, v := range args {
		if v > r {
			r = v
		}
	}
	return r
}
func (this *math) MinInt64(args ...int64) int64 {
	var r int64
	for _, v := range args {
		if v == 0 || v < r {
			r = v
		}
	}
	return r
}