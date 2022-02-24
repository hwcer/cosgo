package updater

func ParseInt(i interface{}) (v int64) {
	switch i.(type) {
	case int32:
		v = int64(i.(int32))
	case int:
		v = int64(i.(int))
	case int64:
		v = i.(int64)
	}
	return
}
