package cache

import "strconv"

const seedDefaultValue = 1000
const datasetKeyBitSize = 36

func Encode(key uint64) string {
	return strconv.FormatUint(key, datasetKeyBitSize)
}

func Decode(key string) (uint64, error) {
	num, err := strconv.ParseUint(key, datasetKeyBitSize, 64)
	if err != nil {
		return 0, err
	} else {
		return num, nil
	}
}
