package utils

func IndexOf(arr []int, tar int) int {
	for k, v := range arr {
		if v == tar {
			return k
		}
	}
	return -1
}
