package session

var listener []func(*Data)

func emit(d *Data) {
	if len(listener) == 0 {
		return
	}
	for _, l := range listener {
		l(d)
	}
}

func On(l func(*Data)) {
	listener = append(listener, l)
}

func Listen(l func(*Data)) {
	listener = append(listener, l)
}
