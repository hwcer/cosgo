package utils

import "strconv"

func StringSliceIndexOf(s []string, tar string) int {
	for i, v := range s {
		if v == tar {
			return i
		}
	}
	return -1
}

func Int32SliceIndexOf(s []int32, tar int32) int {
	for i, v := range s {
		if v == tar {
			return i
		}
	}
	return -1
}

func SliceStringToInt32(s []string) []int32 {
	ret := make([]int32,0)
	for _, v := range s {
		in,err := strconv.Atoi(v)
		if err != nil {
			return nil
		}
		ret = append(ret, int32(in))
	}
	return ret
}